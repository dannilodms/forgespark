import { useQuery } from "@tanstack/react-query"
import { useNavigate } from "@tanstack/react-router"
import { LoaderCircleIcon, LogOutIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { Button } from "#ui/components/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "#ui/components/dropdown-menu"

import { ImageOrInitials } from "../../shared/components/image-or-initials"
import { ErrorAlert } from "../../shared/errors/error-alert"
import { loggedInUserQuery, useLogOut } from "../data/session"

/**
 * The menu of the user logged in, at the end of the header of the shell: its
 * button is their photo, or their initials. It shows their name and their
 * e-mail, then the actions on the session, of which there is one: logging
 * out, which leads to the login screen with its notice. The menu stays open
 * while the logout is sent, and shows its failure.
 */
export function UserMenu() {
  const { t } = useTranslation("access")
  const navigate = useNavigate()
  const { data: user } = useQuery(loggedInUserQuery)
  const logOut = useLogOut()

  if (!user) {
    return null
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger
        render={
          <Button
            variant="ghost"
            size="icon"
            className="rounded-full"
            aria-label={t("userMenu.open", { name: user.name })}
          />
        }
      >
        <ImageOrInitials name={user.name} imageURL={user.photo_url} />
      </DropdownMenuTrigger>
      <DropdownMenuContent
        align="end"
        className="w-auto max-w-[calc(100vw-2rem)] min-w-64"
      >
        <DropdownMenuGroup>
          <DropdownMenuLabel className="flex items-center gap-3 px-1.5 py-2 text-sm font-normal text-popover-foreground">
            <ImageOrInitials
              name={user.name}
              imageURL={user.photo_url}
              size="lg"
            />
            <span className="flex min-w-0 flex-col">
              <span className="truncate font-medium">{user.name}</span>
              <span className="truncate type-caption text-muted-foreground">
                {user.email}
              </span>
            </span>
          </DropdownMenuLabel>
          <DropdownMenuSeparator />
          <DropdownMenuItem
            closeOnClick={false}
            disabled={logOut.isPending}
            onClick={() =>
              logOut.mutate(undefined, {
                onSuccess: () =>
                  void navigate({
                    to: "/login",
                    search: { notice: "logged_out" },
                  }),
              })
            }
          >
            {logOut.isPending ? (
              <LoaderCircleIcon className="animate-spin" />
            ) : (
              <LogOutIcon />
            )}
            {t("userMenu.logOut")}
          </DropdownMenuItem>
        </DropdownMenuGroup>
        {logOut.error ? (
          <div className="max-w-80 p-1.5">
            <ErrorAlert error={logOut.error} />
          </div>
        ) : null}
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
