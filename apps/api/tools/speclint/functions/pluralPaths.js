// A function of the vacuum lint of the OpenAPI spec (../ruleset.yaml). It
// checks Paths in API conventions of CODING_STANDARDS.md: a resource is a
// plural noun, a resource that exists once has a singular path that the
// ruleset lists, and a state change is a POST that ends in the verb of its
// use case.
//
// A segment of a path passes when it is one of these:
//   - part of the prefix every path starts with (`/api`);
//   - a parameter (`{project_id}`);
//   - a plural noun, which here is a word that ends in `s`;
//   - the end of a path the ruleset lists as singular (`/api/setup`);
//   - the last segment of a path whose only method is POST: the verb.

function getSchema() {
  return {
    name: "pluralPaths",
    description: "A resource is a plural noun, unless its path is listed as one that exists once",
    properties: [
      { name: "prefix", description: "The path every operation is under" },
      { name: "singular", description: "The paths of the resources that exist once" },
    ],
  };
}

var methods = ["get", "put", "post", "delete", "options", "head", "patch", "trace"];

function runRule(input) {
  var options = context.ruleAction.functionOptions || {};
  var prefix = options.prefix || "";
  var singular = options.singular || [];
  var results = [];

  Object.keys(input || {}).forEach(function (path) {
    var problem = judge(path, input[path] || {}, prefix, singular);
    if (problem) {
      results.push({ message: problem });
    }
  });
  return results;
}

// judge returns what is wrong with a path, or nothing.
function judge(path, item, prefix, singular) {
  if (path !== prefix && path.indexOf(prefix + "/") !== 0) {
    return "the path " + path + " is not under " + prefix;
  }
  var segments = path.slice(prefix.length).split("/").filter(function (segment) {
    return segment !== "";
  });
  var onlyPost = methods.every(function (method) {
    return method === "post" || item[method] === undefined;
  });

  var walked = prefix;
  for (var index = 0; index < segments.length; index++) {
    var segment = segments[index];
    walked += "/" + segment;
    var last = index === segments.length - 1;
    if (isParameter(segment) || isPlural(segment) || singular.indexOf(walked) >= 0 || (last && onlyPost)) {
      continue;
    }
    return (
      "the path " + path + " has the singular segment `" + segment + "`: " +
      "write the resource in the plural, or list " + walked + " among the singular paths of the ruleset " +
      "when the resource exists once"
    );
  }
  return "";
}

function isParameter(segment) {
  return segment.charAt(0) === "{" && segment.charAt(segment.length - 1) === "}";
}

function isPlural(segment) {
  return segment.charAt(segment.length - 1) === "s";
}
