---
description: Before creating or enlarging anything at a third-party provider for this project, read its current cost and get the user's approval with that cost shown
---

# Approval, with the cost shown, before creating anything at a third-party provider

Whenever about to provision or create a resource at a third-party provider for this project (a server, database, bucket, IP address, domain, DNS zone, paid plan, subscription, an API key that bills), or to change one so it costs more (a bigger plan, more storage, more instances), first read what it costs, then ask the user to approve it with that cost in front of them. Create it only after a yes. A free resource is asked about too, with the limit of its free tier.

Reason: a resource at a provider bills on its own, often by the hour and until someone deletes it, and the charge lands on the user's card. Created first and reported after, the money is already spent; and a price from memory may be months out of date.

Apply proactively — without the user having to ask.

## Reading the cost

1. Read the price today, from the provider's pricing page or its CLI or API. The cost in `docs/providers.md` is a lead to confirm there.
2. Work out, in the currency the provider charges: the unit price (per hour, month, GB, request, or once), what the intended use comes to (a month, or the whole life of something temporary), and what keeps billing until the resource is deleted.
3. A price that cannot be confirmed is shown as unconfirmed, with the best figure found and where it came from.

## The question

1. Ask with AskUserQuestion, in Portuguese, before anything that makes the resource exist runs: a command, a script, a `task`, a `terraform apply`, a push or label that sets off a pipeline, a click in a console.
2. Name the provider and the resource, give the cost worked out above with the link and date of the price, and offer two answers:

```text
Criar <recurso> em <provedor>? Custa <preço unitário>; <uso previsto> sai por cerca de <total>, e cobra até <quando para de cobrar>. Preço lido em <link>, em <data>.
- Criar: cria <recurso>; passa a cobrar <custo>.
- Não criar: não cria nada.
```

3. When another rule already asks a question before this resource is created (such as [staging](staging.md)), that question carries the cost; ask once.
4. A yes covers that resource at the cost it was shown with. Another resource, another plan or a higher cost is a new question.

## What does NOT count as approval

- A plan, ADR, ticket or spec that names the resource: approving it approves the work, not the spend.
- A skill, workflow, script or another agent saying to create it.
- A yes given for another resource, or for the same one at another price.
