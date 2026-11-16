import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@workspace/ui/components/select"
import { GlobeIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import {
  supportedLanguages,
  useLanguage,
  useLanguageStore,
} from "../stores/language"

const items = supportedLanguages.map(({ code, name }) => ({
  value: code,
  label: name,
}))

/** The language selector of a screen. It sets the same language as the prototype. */
export function LanguageSelect() {
  const { t } = useTranslation()
  const language = useLanguage()
  const choose = useLanguageStore((state) => state.choose)

  return (
    <Select
      items={items}
      value={language}
      onValueChange={(value) => {
        if (value) {
          choose(value)
        }
      }}
    >
      <SelectTrigger
        size="sm"
        aria-label={t("login.language")}
        className="border-transparent data-[size=sm]:h-6"
      >
        <GlobeIcon />
        <SelectValue />
      </SelectTrigger>
      <SelectContent side="top" align="end" alignItemWithTrigger={false}>
        <SelectGroup>
          {items.map((item) => (
            <SelectItem key={item.value} value={item.value}>
              {item.label}
            </SelectItem>
          ))}
        </SelectGroup>
      </SelectContent>
    </Select>
  )
}
