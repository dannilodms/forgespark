# Browser tools

## Chrome DevTools MCP

The [Chrome DevTools MCP](https://github.com/ChromeDevTools/chrome-devtools-mcp) server lets an agent open the running app in a real Chrome, record a performance trace, and read the console and the network.

Use it to investigate: a slow screen, a layout shift, an error that only shows in the browser. It is not a test. What must hold on every build is a check in the build (see "Enforcement" in `CODING_STANDARDS.md`).

### Configuration

Claude Code reads `.mcp.json` at the repository root. Other agents do not read that file: register the same command in the agent's own MCP configuration.

```
npx -y chrome-devtools-mcp@1.12.0 --no-usage-statistics --no-performance-crux
```

`.mcp.json` wraps the command in `cmd /c`, the form for native Windows. On another system, call `npx` directly.

- **The version is pinned.** Change the number deliberately, in `.mcp.json` and here; never use `@latest`.
- **`--no-usage-statistics`** turns off the usage data the server sends to Google by default.
- **`--no-performance-crux`** stops performance tools from sending the traced URLs to Google's CrUX API. The app is private and behind login, so there is no field data to fetch.

### Browser profile

The server starts its own Chrome with a profile of its own (`%USERPROFILE%\.cache\chrome-devtools-mcp\chrome-profile` on Windows), separate from the maintainer's browser. The profile is kept between runs, so a login made there survives. Only one browser can use it at a time.

### Recording a trace

1. Start the app with the task runner and open it in the page the server controls.
2. `performance_start_trace` with `reload` set, then `performance_stop_trace`.
3. `performance_analyze_insight` for each insight worth a closer look, such as `LCPBreakdown`.
