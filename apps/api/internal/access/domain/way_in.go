package domain

// WayIn is how a user logs in: with a login code, or, later, through a login
// provider, named by its key. The audit log records it with each login.
type WayIn string

// LoginCodeWayIn is a login with a login code.
const LoginCodeWayIn WayIn = "login_code"
