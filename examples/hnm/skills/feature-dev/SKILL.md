---
name: feature-dev
description: Interrogate a proposed feature before implementation by studying the repository and all project documentation, then asking one product or technical question at a time until intent, scope, acceptance criteria, boundaries, failure modes, and trade-offs are resolved. Use when a user describes a feature idea, asks to clarify or design a feature, wants product and technical requirements challenged, or needs the resulting decisions recorded in docs/prd, docs/adr, and docs/spec.md.
---

# Feature Development

Turn an initial feature idea into an evidence-backed, decision-complete product and technical definition. Call the clarification process “质问”.

## Non-negotiable rules

1. Inspect before asking. Read the code, tests, configuration, and every file under `docs/` before the first question.
2. Ask exactly one question per turn. Wait for the answer before asking the next question.
3. Never hide several decisions inside one compound question or a large form.
4. Ask only what cannot be answered reliably from the repository, documentation, or prior answers.
5. Finish product clarification before technical clarification unless technical discovery exposes a new product ambiguity.
6. Continue until no material uncertainty remains. Do not stop merely because a plausible implementation exists.
7. Do not implement feature code during 质问. Establish and document the contract first.

## 1. Build repository context

- Read applicable `AGENTS.md`, `CLAUDE.md`, repository guidance, and every file under `docs/`, including nested PRDs and ADRs.
- Inspect the repository structure, package manifests, schemas, routes, domain models, public interfaces, tests, and the code paths closest to the proposed feature.
- Trace the current behavior end to end. Prefer direct evidence over assumptions.
- Search for adjacent features, terminology, constraints, TODOs, historical decisions, and existing acceptance tests.
- Inspect current Git changes so the discussion accounts for work already in progress without overwriting unrelated changes.
- Record internally what is known, inferred, contradictory, and still unknown. Do not ask the user to repeat facts already present in the repository.

If existing docs disagree with code, surface the conflict as a concrete question during 质问. Do not silently choose one source as truth.

## 2. Conduct product 质问

Resolve the product contract first. Cover every material area that applies:

- the problem being solved and why it matters;
- target users, actors, and user stories;
- trigger, primary journey, desired outcome, and completion state;
- goals, non-goals, in-scope behavior, and explicitly excluded behavior;
- user-visible states, empty states, errors, permissions, cancellation, and recovery;
- compatibility or migration expectations visible to users;
- the minimum acceptable scope and testable acceptance criteria;
- what would make the feature incomplete or unacceptable.

For each turn:

1. Reconcile the latest answer with code and docs.
2. Identify the single unresolved product decision with the highest impact.
3. Briefly state the relevant evidence or consequence.
4. Ask one focused question and stop.

Prefer an open question when the design space is still unclear. Offer concrete options only when code or prior answers establish a real trade-off. Do not lead the user toward an arbitrary preference.

## 3. Conduct technical 质问

After the product contract is clear, trace the affected implementation paths again and resolve the technical contract. Cover every material area that applies:

- ownership and boundaries between modules, services, clients, and external systems;
- domain model, state transitions, invariants, and source of truth;
- APIs, events, schemas, persistence, migrations, and backward compatibility;
- authentication, authorization, privacy, and trust boundaries;
- concurrency, ordering, idempotency, retries, cancellation, and partial failure;
- performance, scale, availability, observability, rollout, and rollback;
- platform or environment differences;
- validation strategy and the tests needed for acceptance;
- explicit implementation non-goals;
- viable alternatives and their costs, risks, and reversibility.

Do not ask generic architecture questions. First explain what the current code already does, identify the concrete boundary or trade-off, then ask the user to decide the one unresolved point.

When recommending an option, state why it best fits the agreed product contract and existing architecture. Still ask the user to confirm any decision whose trade-off materially changes behavior, scope, risk, cost, or future flexibility.

## 4. Decide when 质问 is complete

Continue the one-question loop until all of the following are true:

- the user story and intended outcome are unambiguous;
- minimum acceptance criteria are observable and testable;
- scope and non-goals are explicit;
- important edge cases and failure behavior are decided;
- affected system boundaries and data ownership are understood;
- material technical alternatives and trade-offs have a decision;
- compatibility, migration, rollout, and validation needs are resolved where relevant;
- no contradiction remains among user answers, code, PRDs, ADRs, and spec;
- no remaining question could materially change the product contract or implementation approach.

Do not ask low-value questions about details that are safely reversible, already conventional in the repository, or discoverable during implementation. Record those as bounded implementation choices instead.

## 5. Write the decisions into docs

Re-read all docs and the relevant code before writing. Follow existing document naming, numbering, and structure. If no convention exists, use concise kebab-case filenames derived from the feature and decision.

### Product requirements

Create or update `docs/prd/<feature>.md` with:

- problem and context;
- target users and user stories;
- goals and non-goals;
- scope and user flow;
- user-visible states and failure behavior;
- minimum acceptance criteria;
- exclusions and resolved product decisions.

Write requirements as testable product statements. Do not include implementation choices unless they are themselves product constraints.

### Architecture decisions

Create or update one document under `docs/adr/` for each material technical decision. Include:

- status;
- context and forces;
- decision;
- considered alternatives;
- trade-offs and consequences;
- boundaries, failure behavior, migration, and validation implications where relevant.

Do not turn the ADR into a code walkthrough. Capture why the decision exists and what it constrains.

### General specification

Update `docs/spec.md` with the stable, implementation-independent rules established by the feature: shared terminology, externally observable contracts, system-wide invariants, or general constraints. Keep feature-specific rationale in the PRD and decision rationale in ADRs. Avoid duplicating whole sections across all three documents.

Every completed 质问 must leave a PRD record, an ADR record, and an appropriate spec update.

## 6. Verify the documentation set

- Review the full documentation diff against the final answers and current code.
- Check that PRD acceptance criteria do not conflict with ADR consequences or `spec.md` invariants.
- Remove unresolved placeholders and chat-like transcripts.
- Preserve unrelated docs and user changes.
- Report the files written, the product contract, the technical decisions, and any intentionally deferred implementation details.

If the user also requested implementation, begin it only after this documentation set is complete. Treat the accepted PRD, ADRs, and spec as the source of truth and flag any new contradiction before changing code.
