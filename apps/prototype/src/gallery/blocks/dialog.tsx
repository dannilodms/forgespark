import { Button } from "@workspace/ui/components/button"
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@workspace/ui/components/dialog"
import {
  Field,
  FieldDescription,
  FieldGroup,
  FieldLabel,
} from "@workspace/ui/components/field"
import { Input } from "@workspace/ui/components/input"
import {
  Sheet,
  SheetClose,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@workspace/ui/components/sheet"
import { StatusIndicator } from "@workspace/ui/components/status-indicator"
import { useTranslation } from "react-i18next"

import { overlays } from "../explanations/overlays"
import { partTitle } from "../explanations/shared"
import { freeText, sampleRequest } from "../free-text"
import { Frame, Part, Section } from "../parts"

const longFields = ["subject", "deadline", "owner", "notes"] as const

export function DialogBlock() {
  const { t } = useTranslation()
  const text = overlays.dialog
  const close = t("gallery.common.close")
  const cancel = t("gallery.common.cancel")

  return (
    <Section {...text.about}>
      <Part title={partTitle.typical}>
        <Frame>
          <Dialog>
            <DialogTrigger render={<Button variant="outline" />}>
              {text.openRename}
            </DialogTrigger>
            <DialogContent closeLabel={close}>
              <DialogHeader>
                <DialogTitle>{t("gallery.dialog.rename.title")}</DialogTitle>
                <DialogDescription>
                  {t("gallery.dialog.rename.description")}
                </DialogDescription>
              </DialogHeader>
              <FieldGroup>
                <Field>
                  <FieldLabel htmlFor="dialog-project">
                    {t("gallery.common.projectName")}
                  </FieldLabel>
                  <Input id="dialog-project" defaultValue={freeText.project} />
                </Field>
              </FieldGroup>
              <DialogFooter>
                <DialogClose render={<Button variant="ghost" />}>
                  {cancel}
                </DialogClose>
                <DialogClose render={<Button />}>
                  {t("gallery.common.saveChanges")}
                </DialogClose>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        </Frame>
      </Part>

      <Part {...text.confirming}>
        <Frame>
          <Dialog>
            <DialogTrigger render={<Button variant="outline" />}>
              {text.openDelete}
            </DialogTrigger>
            <DialogContent showCloseButton={false}>
              <DialogHeader>
                <DialogTitle>
                  {t("gallery.dialog.delete.title", { name: freeText.project })}
                </DialogTitle>
                <DialogDescription>
                  {t("gallery.dialog.delete.description", { count: 7 })}
                </DialogDescription>
              </DialogHeader>
              <DialogFooter>
                <DialogClose render={<Button variant="ghost" />}>
                  {cancel}
                </DialogClose>
                <DialogClose render={<Button variant="destructive" />}>
                  {t("gallery.common.deleteProject")}
                </DialogClose>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        </Frame>
      </Part>

      <Part {...text.edgeCases}>
        <Frame>
          <Dialog>
            <DialogTrigger render={<Button variant="outline" />}>
              {text.openLong}
            </DialogTrigger>
            <DialogContent closeLabel={close}>
              <DialogHeader>
                <DialogTitle>{t("gallery.dialog.long.title")}</DialogTitle>
                <DialogDescription>
                  {t("gallery.dialog.long.description")}
                </DialogDescription>
              </DialogHeader>
              <FieldGroup>
                {longFields.map((field) => (
                  <Field key={field}>
                    <FieldLabel htmlFor={`dialog-long-${field}`}>
                      {t(`gallery.dialog.long.fields.${field}`)}
                    </FieldLabel>
                    <Input id={`dialog-long-${field}`} />
                    <FieldDescription>
                      {t("gallery.dialog.long.hint")}
                    </FieldDescription>
                  </Field>
                ))}
              </FieldGroup>
              <DialogFooter>
                <DialogClose render={<Button variant="ghost" />}>
                  {cancel}
                </DialogClose>
                <DialogClose render={<Button />}>
                  {t("gallery.dialog.long.submit")}
                </DialogClose>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        </Frame>
      </Part>
    </Section>
  )
}

const sides = ["left", "top", "bottom"] as const

export function SheetBlock() {
  const { t } = useTranslation()
  const text = overlays.sheet
  const close = t("gallery.common.close")

  return (
    <Section {...text.about}>
      <Part title={partTitle.typical}>
        <Frame>
          <Sheet>
            <SheetTrigger render={<Button variant="outline" />}>
              {text.openRequest}
            </SheetTrigger>
            <SheetContent closeLabel={close}>
              <SheetHeader>
                <SheetTitle>{sampleRequest.title}</SheetTitle>
                <SheetDescription className="font-mono">
                  {sampleRequest.id}
                </SheetDescription>
              </SheetHeader>
              <div className="flex flex-col gap-3 px-4">
                <StatusIndicator status={sampleRequest.status}>
                  {t(`gallery.common.status.${sampleRequest.status}`)}
                </StatusIndicator>
                <p>{freeText.description}</p>
              </div>
              <SheetFooter>
                <Button>{t("gallery.sheet.openRequest")}</Button>
                <SheetClose render={<Button variant="ghost" />}>
                  {close}
                </SheetClose>
              </SheetFooter>
            </SheetContent>
          </Sheet>
        </Frame>
      </Part>

      <Part {...text.sides}>
        <Frame>
          {sides.map((side) => (
            <Sheet key={side}>
              <SheetTrigger render={<Button variant="outline" />}>
                {text.side[side].trigger}
              </SheetTrigger>
              <SheetContent side={side} closeLabel={close}>
                <SheetHeader>
                  <SheetTitle>{t("gallery.sheet.filters")}</SheetTitle>
                  <SheetDescription>{text.side[side].content}</SheetDescription>
                </SheetHeader>
                <SheetFooter>
                  <SheetClose render={<Button variant="secondary" />}>
                    {close}
                  </SheetClose>
                </SheetFooter>
              </SheetContent>
            </Sheet>
          ))}
        </Frame>
      </Part>

      <Part {...text.edgeCases}>
        <Frame>
          <Sheet>
            <SheetTrigger render={<Button variant="outline" />}>
              {text.openLong}
            </SheetTrigger>
            <SheetContent closeLabel={close}>
              <SheetHeader>
                <SheetTitle>{freeText.longTitle}</SheetTitle>
                <SheetDescription>
                  {t("gallery.sheet.openedBy")}
                </SheetDescription>
              </SheetHeader>
              <div className="flex min-h-0 flex-1 flex-col gap-3 overflow-y-auto px-4">
                {[1, 2, 3, 4, 5, 6].map((paragraph) => (
                  <p key={paragraph}>{freeText.longParagraph}</p>
                ))}
              </div>
              <SheetFooter>
                <SheetClose render={<Button variant="secondary" />}>
                  {close}
                </SheetClose>
              </SheetFooter>
            </SheetContent>
          </Sheet>
        </Frame>
      </Part>
    </Section>
  )
}
