package restapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"
	"slices"
	"strings"

	"github.com/danielgtaylor/huma/v2"
)

// Three failures never reach the functions Huma builds its errors with: a
// route nobody registered, a method a route does not take, and a panic. Each
// has a handler here, so that they answer in the error format too.

// Prefix is the path everything the binary serves to the outside is under.
const Prefix = "/api"

// methods are the ones a route may take, in the order the `Allow` header
// lists them.
var methods = []string{
	http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete,
}

// unrouted answers what no operation takes, under the prefix: 405 when the
// path is a route under another method, 404 otherwise.
type unrouted struct {
	mux *http.ServeMux
}

// mountUnrouted registers the handler on every path under the prefix. The
// mux prefers the pattern of an operation, which names a method and a longer
// path, so this one receives only what matched nothing else.
func mountUnrouted(mux *http.ServeMux) {
	handler := unrouted{mux: mux}
	mux.Handle(Prefix, handler)
	mux.Handle(Prefix+"/", handler)
}

func (u unrouted) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	allowed := slices.DeleteFunc(slices.Clone(methods), func(method string) bool { return !u.routes(r, method) })
	if len(allowed) == 0 {
		writeProblem(w, newProblem(r.Context(), http.StatusNotFound, codeOfStatus(http.StatusNotFound), ""))
		return
	}
	w.Header().Set("Allow", strings.Join(allowed, ", "))
	writeProblem(w, newProblem(r.Context(), http.StatusMethodNotAllowed, codeOfStatus(http.StatusMethodNotAllowed), ""))
}

// routes tells whether the path of the request is a route under this method.
func (u unrouted) routes(r *http.Request, method string) bool {
	asked := r.Clone(r.Context())
	asked.Method = method
	_, pattern := u.mux.Handler(asked)
	return pattern != "" && pattern != Prefix && pattern != Prefix+"/"
}

// writeProblem answers with a problem outside Huma.
func writeProblem(w http.ResponseWriter, problem *Problem) {
	w.Header().Set("Content-Type", problemContentType)
	w.WriteHeader(problem.Status)
	// The caller that does not read its answer is not a failure of the API.
	_ = json.NewEncoder(w).Encode(problem)
}

// recoverPanic answers a panic of an operation with a 500 in the error
// format. The value of the panic and its stack go to the logs, as the raw
// error of the request, and never to the answer.
func (a *API) recoverPanic(ctx huma.Context, next func(huma.Context)) {
	defer func() {
		panicked := recover()
		if panicked == nil {
			return
		}
		if panicked == http.ErrAbortHandler {
			// The handler asked the server to drop the connection.
			panic(panicked)
		}
		// Nothing more can be done for a caller that is gone.
		_ = huma.WriteErr(a.API, ctx, http.StatusInternalServerError, "panic", fmt.Errorf("%v\n%s", panicked, debug.Stack()))
	}()
	next(ctx)
}
