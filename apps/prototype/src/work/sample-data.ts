/*
 * SAMPLE DATA. Nothing here is real and nothing here defines the product:
 * the people, the clients, the items and their statuses exist only to fill
 * the sample work screen, so the style can be judged on a full work surface.
 * The product's own entities and vocabulary are in GLOSSARY.md.
 *
 * Titles and descriptions stand for text typed by a person, so they stay in
 * one language. Every other text of the screen comes from the translation
 * files.
 */
import type { Status } from "@workspace/ui/components/status-indicator"

import type { ActivityEntry, Place, WorkItem } from "./work-items"

/** The sample person who is logged in. */
export const samplePerson = {
  name: "Ana Lima",
  email: "ana.lima@example.com",
}

const bruno = "Bruno Tavares"
const carla = "Carla Mendes"

/** The sample clients a new project can belong to. */
export const sampleClients = [
  "Northwind Traders",
  "Acme Logistics",
  "Vale Foods",
]

/** One sample item: opened by one person, then moved to its status by another. */
function sample(
  place: Place,
  id: string,
  title: string,
  status: Status,
  owner: string,
  group: string,
  opened: string,
  updated: string,
  description: string
): WorkItem {
  const activity: ActivityEntry[] = [
    { kind: "opened", at: opened, who: samplePerson.name },
  ]
  if (status !== "working") {
    activity.unshift({ kind: "statusChanged", status, at: updated, who: owner })
  }
  return {
    id,
    place,
    title,
    status,
    owner,
    group,
    updated,
    description,
    activity,
  }
}

/** The sample items of both places, the most recent first. */
export const sampleItems: WorkItem[] = [
  sample(
    "requests",
    "REQ-0142",
    "Export leaves out archived projects",
    "working",
    bruno,
    "Billing portal",
    "2026-11-09",
    "2026-11-13",
    "The client asked for the export to include archived projects and the files attached to each request."
  ),
  sample(
    "requests",
    "REQ-0141",
    "Invoice total is rounded twice",
    "error",
    carla,
    "Billing portal",
    "2026-11-06",
    "2026-11-12",
    "The total of an invoice with more than one tax is one cent off. The fix was deployed and had to be rolled back."
  ),
  sample(
    "requests",
    "REQ-0139",
    "New logo on the invoice",
    "attention",
    samplePerson.name,
    "Billing portal",
    "2026-11-03",
    "2026-11-10",
    "Waiting for the client to send the logo as a vector file."
  ),
  sample(
    "requests",
    "REQ-0138",
    "Push notification when a delivery is late",
    "working",
    carla,
    "Delivery tracking app",
    "2026-11-02",
    "2026-11-10",
    "The driver and the customer are both told when a delivery is more than 15 minutes late."
  ),
  sample(
    "requests",
    "REQ-0136",
    "Map does not load on a slow connection",
    "attention",
    bruno,
    "Delivery tracking app",
    "2026-10-28",
    "2026-11-07",
    "It could not be reproduced. The client was asked for the model of the phone and the name of the carrier."
  ),
  sample(
    "requests",
    "REQ-0131",
    "Login by e-mail code",
    "done",
    samplePerson.name,
    "Customer portal",
    "2026-10-20",
    "2026-11-05",
    "A person without a password receives a six-digit code by e-mail."
  ),
  sample(
    "requests",
    "REQ-0129",
    "Report of hours by project and by month",
    "done",
    bruno,
    "Intranet",
    "2026-10-19",
    "2026-11-04",
    ""
  ),
  sample(
    "requests",
    "REQ-0127",
    "Nightly import of payments",
    "error",
    carla,
    "Billing portal",
    "2026-10-15",
    "2026-11-03",
    "The import stops at the first payment with no document number. The bank changed the layout of its file."
  ),
  sample(
    "requests",
    "REQ-0125",
    "Search by client name ignores accents",
    "done",
    samplePerson.name,
    "Customer portal",
    "2026-10-13",
    "2026-10-30",
    "Searching for Sao Paulo finds São Paulo."
  ),
  sample(
    "requests",
    "REQ-0122",
    "Dark theme for the driver app",
    "working",
    bruno,
    "Delivery tracking app",
    "2026-10-09",
    "2026-10-29",
    "Drivers asked for it for the night shift."
  ),
  sample(
    "requests",
    "REQ-0118",
    "Export of the delivery history to a spreadsheet, with one row per stop and the signature of who received it",
    "attention",
    carla,
    "Delivery tracking app",
    "2026-10-06",
    "2026-10-27",
    "The estimate was sent, and the client has not approved it yet."
  ),
  sample(
    "requests",
    "REQ-0113",
    "Two-step approval for a refund",
    "done",
    samplePerson.name,
    "Billing portal",
    "2026-09-29",
    "2026-10-22",
    "A refund above the limit of the operator needs the approval of a manager."
  ),
  sample(
    "projects",
    "PRJ-0012",
    "Billing portal",
    "working",
    samplePerson.name,
    "Northwind Traders",
    "2026-03-02",
    "2026-11-13",
    "Invoices, payments and refunds of the client, for its finance team."
  ),
  sample(
    "projects",
    "PRJ-0011",
    "Delivery tracking app",
    "attention",
    bruno,
    "Acme Logistics",
    "2026-05-11",
    "2026-11-10",
    "The app of the drivers and the page the customer follows a delivery on. The delivery date of the next version depends on an approval."
  ),
  sample(
    "projects",
    "PRJ-0009",
    "Customer portal",
    "done",
    carla,
    "Northwind Traders",
    "2025-09-15",
    "2026-11-05",
    "Where the customers of the client see their orders. Delivered, and in maintenance."
  ),
  sample(
    "projects",
    "PRJ-0008",
    "Intranet",
    "working",
    samplePerson.name,
    "Vale Foods",
    "2026-01-19",
    "2026-11-04",
    ""
  ),
  sample(
    "projects",
    "PRJ-0005",
    "Legacy data migration",
    "error",
    bruno,
    "Acme Logistics",
    "2025-11-03",
    "2026-10-21",
    "Stopped: the old system has no export, and the access to its database was not granted."
  ),
]
