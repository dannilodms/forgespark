import { StrictMode } from "react"
import { createRoot } from "react-dom/client"

import { createApp } from "./app"
import { startTheme } from "./shared/theme"
import "./ui/styles/globals.css"

startTheme()

const App = createApp({ languages: navigator.languages })

const root = document.getElementById("root")
if (!root) {
  throw new Error("index.html has no #root element")
}

createRoot(root).render(
  <StrictMode>
    <App />
  </StrictMode>
)
