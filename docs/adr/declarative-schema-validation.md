# Validate ctxl schemas declaratively

Status: Accepted

## Context

Schema validation must not duplicate required fields, patterns, enums, or conditional combinations in handwritten Go code. Doing so would make executable code, rather than the machine-readable contract, the effective definition of valid schema input.

## Decision

- Check in a canonical ctxl JSON Schema using JSON Schema Draft 2020-12.
- Compile and evaluate it with `github.com/santhosh-tekuri/jsonschema/v6` before decoding into runtime Go types.
- Generate the Go JSON decoder and its basic validation from that same document with `go-jsonschema`; do not maintain validator tags on runtime types.
- Reject unknown properties unless an explicit extension point says otherwise.
- Express required values, patterns, enums, conditional requirements, mutually exclusive values, paths, versions, and configuration shapes in the JSON Schema.
- Return stable, structured validation errors carrying the source file, JSON instance path, rule, and human-readable message.
- Apply the same validation entrypoint from code generation and the retained public schema parsing API.
- Let generated decoding apply literal JSON Schema defaults. After validation, keep only cross-property derivation that JSON Schema defaults cannot express, such as deriving CLI and store names from the root name; derivation is not a second validator.

## Alternatives considered

### Extend the handwritten `Schema.Validate`

Rejected. It duplicates a standard schema language, produces inconsistent error locations, and will grow with each generator option.

### Validate only decoded structs with handwritten Go tags

Rejected. Handwritten tags duplicate the contract and do not by themselves define all Draft 2020-12 behavior or provide a reusable JSON contract for editors and other tools.

### Generate a JSON Schema from Go structs

Rejected as the source of truth. Reflection can describe field shapes but does not naturally own all cross-field rules and generator failure boundaries. The checked-in JSON Schema is the contract; Go types adapt to it.

## Consequences

- Schema authors receive standard JSON paths and rule failures and can use the checked-in schema in editors.
- Generated Go decoding changes with the JSON Schema. Contract fixtures prove accepted and rejected inputs at the canonical boundary.
- The new parser is intentionally stricter than the current decoder, including unknown-property rejection.
- The generator validates completely before touching output.

## Validation

- Run the upstream JSON Schema test behavior exercised by the selected library through focused integration fixtures.
- Add positive fixtures for defaults, both output modes, identity overrides, implicit and explicit built-in Skills, multiple custom Skills, and built-in injection options.
- Add negative fixtures for unknown keys, invalid enum combinations, extra custom-Skill fields, multiple built-in entries, missing directories, unsafe paths, and conflicting effective Skill names.
