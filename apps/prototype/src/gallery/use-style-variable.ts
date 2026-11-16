import { useEffect, useState } from "react"

/** The value a style variable has right now. It is read again when the theme changes. */
export function useStyleVariable(variable: string) {
  const [value, setValue] = useState("")

  useEffect(() => {
    const read = () => {
      const style = getComputedStyle(document.documentElement)
      setValue(style.getPropertyValue(`--${variable}`).trim())
    }
    read()
    const observer = new MutationObserver(read)
    observer.observe(document.documentElement, { attributes: true })
    return () => observer.disconnect()
  }, [variable])

  return value
}
