import { GlobeIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "#ui/components/select"

import { useLanguage } from "../locales/i18n"
import { type Language, supportedLanguages } from "../locales/languages"

const items = supportedLanguages.map(({ code, name }) => ({
  value: code,
  label: name,
}))

/** The language selector of the status bar. It changes every text at once. */
export function LanguageSelect() {
  const { t } = useTranslation()
  const { language, choose } = useLanguage()

  return (
    <Select<Language>
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
        aria-label={t("language")}
        className="border-transparent data-[size=sm]:h-7"
      >
        <GlobeIcon />
        <SelectValue />
      </SelectTrigger>
      <SelectContent side="top" align="end" alignItemWithTrigger={false}>
        <SelectGroup>
          {items.map((item) => (
            <SelectItem key={item.value} value={item.value} lang={item.value}>
              {item.label}
            </SelectItem>
          ))}
        </SelectGroup>
      </SelectContent>
    </Select>
  )
}
