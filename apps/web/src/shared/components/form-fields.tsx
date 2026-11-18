import type { InputHTMLAttributes, ReactNode } from "react"
import { useTranslation } from "react-i18next"

import {
  Combobox,
  ComboboxContent,
  ComboboxEmpty,
  ComboboxInput,
  ComboboxItem,
  ComboboxList,
} from "#ui/components/combobox"
import { Field, FieldDescription, FieldLabel } from "#ui/components/field"
import { Input } from "#ui/components/input"

import { FieldProblemMessage } from "../errors/field-problem"
import type { FieldProblem } from "../errors/problems"
import { useLanguage } from "../locales/i18n"
import type { ListItem } from "../places/places"

/**
 * A field of TanStack Form that holds a text, as far as the components below
 * read it. `form.Field` hands one to its children.
 */
type TextFieldApi = {
  name: string
  state: { value: string; meta: { errors: unknown[] } }
  handleChange: (value: string) => void
  handleBlur: () => void
}

type FieldTexts = {
  /** The label, which is also the name of the control. */
  label: string
  /** Whether the field may be left empty: its label says so. */
  optional?: boolean
  /** A sentence under the label, such as the form the value takes. */
  description?: string
}

/**
 * The parts every field shares: its label, its description, what is wrong
 * with it, and the attributes its control takes from them. The control is
 * named after the place of the field in the body, so that a form can find
 * the first field with a problem (`focusTheFirstProblem`).
 */
function useFieldParts(
  field: TextFieldApi,
  { label, optional, description }: FieldTexts
) {
  const { t } = useTranslation()
  const id = `field-${field.name.replaceAll(".", "-")}`
  const problem = field.state.meta.errors[0] as FieldProblem | undefined
  const describedBy = [
    description ? `${id}-description` : "",
    problem ? `${id}-problem` : "",
  ].filter(Boolean)

  return {
    problem,
    control: {
      id,
      "data-field": field.name,
      "aria-invalid": problem ? true : undefined,
      "aria-required": optional ? undefined : true,
      "aria-describedby":
        describedBy.length > 0 ? describedBy.join(" ") : undefined,
      onBlur: field.handleBlur,
    },
    wrap: (control: ReactNode) => (
      <Field data-invalid={problem ? true : undefined}>
        <FieldLabel htmlFor={id}>
          {optional ? t("form.optional", { label }) : label}
        </FieldLabel>
        {control}
        {description ? (
          <FieldDescription id={`${id}-description`}>
            {description}
          </FieldDescription>
        ) : null}
        {problem ? (
          <FieldProblemMessage id={`${id}-problem`} problem={problem} />
        ) : null}
      </Field>
    ),
  }
}

/** A field of text: a label, the input, and what is wrong with it. */
export function TextField({
  field,
  label,
  optional,
  description,
  ...input
}: { field: TextFieldApi } & FieldTexts &
  Pick<
    InputHTMLAttributes<HTMLInputElement>,
    "type" | "autoComplete" | "inputMode" | "placeholder" | "spellCheck"
  >) {
  const { control, wrap } = useFieldParts(field, {
    label,
    optional,
    description,
  })

  return wrap(
    <Input
      {...input}
      {...control}
      value={field.state.value}
      onChange={(event) => field.handleChange(event.target.value)}
    />
  )
}

/**
 * A field whose value is chosen from a list that is filtered by typing: a
 * country, a currency, a time zone. The field holds the code of the item;
 * the screen shows its name, and filters by it in the language of the
 * screen, whatever the case and the accents.
 */
export function ListField({
  field,
  items,
  label,
  optional,
  description,
}: { field: TextFieldApi; items: ListItem[] } & FieldTexts) {
  const { t } = useTranslation()
  const { language } = useLanguage()
  const { control, wrap } = useFieldParts(field, {
    label,
    optional,
    description,
  })
  const chosen = items.find((item) => item.value === field.state.value) ?? null

  return wrap(
    <Combobox
      items={items}
      value={chosen}
      onValueChange={(item: ListItem | null) =>
        field.handleChange(item?.value ?? "")
      }
      isItemEqualToValue={(item: ListItem, value: ListItem) =>
        item.value === value.value
      }
      locale={language}
      autoHighlight
    >
      <ComboboxInput
        {...control}
        triggerLabel={t("form.showTheList", { label })}
      />
      <ComboboxContent>
        <ComboboxEmpty>{t("form.nothingMatches")}</ComboboxEmpty>
        <ComboboxList>
          {(item: ListItem) => (
            <ComboboxItem key={item.value} value={item}>
              {item.label}
            </ComboboxItem>
          )}
        </ComboboxList>
      </ComboboxContent>
    </Combobox>
  )
}

/**
 * Puts the focus on the first control, in the order of the screen, whose
 * field has a problem, inside `container`: after a refused submission, or
 * when a step opens on a field the API refused. Tells whether there was one.
 */
export function focusTheFirstProblem(
  container: HTMLElement | null,
  fieldsWithProblems: readonly string[]
) {
  const controls = container?.querySelectorAll<HTMLElement>("[data-field]")
  const first = [...(controls ?? [])].find((control) =>
    fieldsWithProblems.includes(control.dataset.field ?? "")
  )
  first?.focus()
  return first !== undefined
}
