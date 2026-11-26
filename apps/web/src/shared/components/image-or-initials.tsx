import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"

import { Avatar, AvatarFallback, AvatarImage } from "#ui/components/avatar"
import { cn } from "#ui/lib/utils"

/** A letter of any script, which a word must hold to give an initial. */
const letter = /\p{L}/u

/** The characters as a person sees them: a letter with its accent is one. */
const graphemes = new Intl.Segmenter(undefined, { granularity: "grapheme" })

/**
 * The words that end a name to tell a generation apart, in the languages of
 * the system, written in lowercase and without a full stop.
 */
const generations = new Set([
  "jr",
  "sr",
  "ii",
  "iii",
  "iv",
  "júnior",
  "junior",
  "filho",
  "filha",
  "neto",
  "neta",
  "sobrinho",
  "sobrinha",
])

/**
 * The first character of a text that passes `test`, with its accent: by
 * default, its first character.
 */
function firstCharacterOf(
  text: string,
  test: (character: string) => boolean = () => true
) {
  for (const { segment } of graphemes.segment(text)) {
    if (test(segment)) {
      return segment
    }
  }
  return ""
}

/**
 * The initials of a name: the first letter of its first word and of its
 * last, in capitals, in the language of the screen. A word with no letter,
 * such as an emoji, gives none, and a generation that ends a name of three
 * words or more is left out: "João da Silva Filho" is JS. One word gives one
 * letter; a name with no letter gives its first character.
 */
function initialsOf(name: string, language: string) {
  const words = name
    .trim()
    .split(/\s+/)
    .filter((word) => letter.test(word))
  const isGeneration = (word = "") =>
    generations.has(word.toLocaleLowerCase(language).replace(/\.$/, ""))
  while (words.length > 2 && isGeneration(words.at(-1))) {
    words.pop()
  }
  if (words.length === 0) {
    return firstCharacterOf(name.trim())
  }
  const isLetter = (character: string) => letter.test(character)
  const first = words[0]
  const last = words.length > 1 ? words.at(-1) : undefined
  return [first, last]
    .map((word) => (word ? firstCharacterOf(word, isLetter) : ""))
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
 * read. A person is a circle. The consultancy, whose logo may have any shape,
 * is a box with rounded corners where the whole logo shows: as high as the
 * circle, and as wide as the logo up to three times that, so a wide logo
 * stays legible; its initials are a square. It is decorative: the name is
 * written beside it, or names the control that holds it.
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
      className={cn(
        square && "w-fit max-w-24 min-w-8 rounded-md after:rounded-md",
        className
      )}
    >
      {shown ? (
        <AvatarImage
          src={shown}
          alt=""
          className={cn(
            square &&
              "aspect-auto h-full w-auto max-w-full rounded-md bg-card object-contain"
          )}
        />
      ) : null}
      <AvatarFallback
        className={cn(square && "aspect-square w-auto rounded-md")}
      >
        {initialsOf(name, i18n.language)}
      </AvatarFallback>
    </Avatar>
  )
}
