/**
 * SAMPLE DATA of the gallery: what a person would have typed. Names, titles
 * and descriptions are free text, so they stay in the language they were
 * written in, whatever the language of the prototype (ADR-0005). Every other
 * sample text of a block comes from the translation files.
 *
 * The long ones were typed in Portuguese, which runs longer than English.
 * Nothing here is real, and nothing here defines the product.
 */
export const freeText = {
  product: "Software Factory",
  project: "Billing portal",
  projectFile: "billing-portal.zip",
  client: "Northwind Traders",
  clientAddress: "northwind.example",

  ana: "Ana Lima",
  anaInitials: "AL",
  bruno: "Bruno Tavares",
  brunoInitials: "BT",
  carla: "Carla Mendes",
  carlaInitials: "CM",
  longName: "Maria Fernanda de Albuquerque Cavalcanti e Vasconcelos",
  longNameInitials: "MF",
  unfinishedEmail: "ana@example",

  description:
    "The client asked for the export to include archived projects and the files attached to each request.",
  shortDescription: "The export leaves out archived projects.",
  note: "Billing closes on the last working day.",
  search: "archived export",
  searchWithNoMatch: "payroll",

  longProject: "Portal de cobrança e de acompanhamento das solicitações",
  otherLongProject: "Aplicativo de acompanhamento das entregas",
  shortProject: "Portal de cobrança",
  longClient: "Comércio e Distribuição de Alimentos Vale do Rio Doce Ltda.",
  longTitle:
    "A exportação das solicitações não inclui os projetos arquivados nem os anexos de cada solicitação",
  longDescription:
    "O cliente pediu que a exportação inclua os projetos arquivados e os anexos de cada solicitação: https://clientes.exemplo.com.br/projetos/portal-de-cobranca/solicitacoes/exportacao",
  longParagraph:
    "O cliente pediu que a exportação inclua os projetos arquivados e os anexos de cada solicitação, com o nome de quem a abriu e a data de cada alteração.",
  longAddress:
    "https://clientes.exemplo.com.br/projetos/portal-de-cobranca/solicitacoes",
  longHost: "clientes.exemplo.com.br/",
  longPath: "portal-de-cobranca",
} as const

/** The sample requests, one for each status. */
export const sampleRequests = [
  {
    id: "REQ-0142",
    title: "Export leaves out archived projects",
    client: "Northwind Traders",
    status: "working",
    hours: 6,
  },
  {
    id: "REQ-0139",
    title: "New logo on the invoice",
    client: "Northwind Traders",
    status: "attention",
    hours: 1.5,
  },
  {
    id: "REQ-0131",
    title: "Login by e-mail code",
    client: "Contoso",
    status: "done",
    hours: 24,
  },
  {
    id: "REQ-0127",
    title: "Nightly import of payments",
    client: "Fabrikam",
    status: "error",
    hours: 11,
  },
] as const

/** The request the blocks show most: the first of the sample. */
export const sampleRequest = sampleRequests[0]
