import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"

import { Avatar, AvatarFallback, AvatarImage } from "#ui/components/avatar"
import { cn } from "#ui/lib/utils"

/**
 * The initials of a name: the first letter of its first word and of its
 * last, in capitals, in the language of the screen. One word gives one
 * letter.
 */
function initialsOf(name: string, language: string) {
  const words = name.trim().split(/\s+/).filter(Boolean)
  const first = words.at(0)
  const last = words.length > 1 ? words.at(-1) : undefined
  return [first, last]
    .map((word) => (word ? Array.from(word)[0] : ""))
    .join("")
    .toLocaleUpperCase(language)
}

/**
 * The URL to show an image from: the newest one given, once the browser holds
 * its picture, and the one shown before while it loads. The API signs a new
 * URL for the same image every few minutes, and the image does not blink to
 * the initials each time: the picture of the new URL is already loaded when
 * it takes the place of the old one.
 */
function useSteadyImageURL(url: string | undefined) {
  const [shown, setShown] = useState(url)

  useEffect(() => {
    if (!url || url === shown) {
      return
    }
    const image = new Image()
    const show = () => setShown(url)
    image.addEventListener("load", show)
    image.addEventListener("error", show)
    image.src = url
    return () => {
      image.removeEventListener("load", show)
      image.removeEventListener("error", show)
    }
  }, [url, shown])

  return url ? (shown ?? url) : undefined
}

/**
 * The image of a person or of the consultancy, read from the file storage,
 * and its initials while there is none, while it loads or when it cannot be
 * read. A person is a circle; the consultancy, whose logo may have any shape,
 * a square with rounded corners, where the whole logo shows. It is
 * decorative: the name is written beside it, or names the control that holds
 * it.
 */
export function ImageOrInitials({
  name,
  imageURL,
  shape = "circle",
  size = "default",
  className,
}: {
  name: string
  imageURL?: string
  shape?: "circle" | "square"
  size?: "default" | "lg"
  className?: string
}) {
  const { i18n } = useTranslation()
  const shown = useSteadyImageURL(imageURL)
  const square = shape === "square"

  return (
    <Avatar
      size={size}
      aria-hidden
      className={cn(square && "rounded-md after:rounded-md", className)}
    >
      {shown ? (
        <AvatarImage
          src={shown}
          alt=""
          className={cn(square && "rounded-md bg-card object-contain")}
        />
      ) : null}
      <AvatarFallback className={cn(square && "rounded-md")}>
        {initialsOf(name, i18n.language)}
      </AvatarFallback>
    </Avatar>
  )
}
