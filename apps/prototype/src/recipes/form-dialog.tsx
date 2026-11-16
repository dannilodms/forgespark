import { Button } from "@workspace/ui/components/button"
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@workspace/ui/components/dialog"
import { FieldError, FieldGroup } from "@workspace/ui/components/field"
import { CircleAlertIcon } from "lucide-react"
import type { ComponentProps, ReactNode } from "react"

/**
 * The validation error of a field of the form dialog, under the field it is
 * about: the error colour, a sign and a sentence that says what to do.
 */
export function FormDialogError({
  children,
  ...props
}: ComponentProps<typeof FieldError>) {
  return (
    <FieldError className="flex items-start gap-1" {...props}>
      <CircleAlertIcon className="mt-0.5 size-3.5 shrink-0" />
      {children}
    </FieldError>
  )
}

/**
 * The form dialog: a dialog with fields, one primary action and a cancel
 * action. It reports a validation error on the field.
 *
 * The children are the fields. Enter in a field submits, Escape and the
 * cancel action close the dialog and keep nothing. The dialog stays open
 * when `onSubmit` finds an error, and the screen then marks the field.
 */
export function FormDialog({
  open,
  onOpenChange,
  title,
  description,
  submitLabel,
  cancelLabel,
  onSubmit,
  children,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  title: string
  description: string
  submitLabel: string
  cancelLabel: string
  onSubmit: () => void
  children: ReactNode
}) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <form
          noValidate
          className="grid gap-4"
          onSubmit={(event) => {
            event.preventDefault()
            onSubmit()
          }}
        >
          <DialogHeader>
            <DialogTitle>{title}</DialogTitle>
            <DialogDescription>{description}</DialogDescription>
          </DialogHeader>
          <FieldGroup>{children}</FieldGroup>
          <DialogFooter>
            <DialogClose render={<Button type="button" variant="ghost" />}>
              {cancelLabel}
            </DialogClose>
            <Button type="submit">{submitLabel}</Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
