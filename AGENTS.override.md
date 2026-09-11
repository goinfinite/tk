# AGENTS.override.md

## Value Object Inputs

Infra utility methods MUST accept `tkValueObject` types when a matching value object exists.
Older methods that take primitive types MUST migrate as they are touched.

## Layer Documentation

`src/domain/README.md`, `src/infra/README.md`, and `src/presentation/README.md` document each layer's components. `SKILL.md` routes consumer agents to them.
When you add, remove, rename, or change the behavior of a component, update the matching layer README in the same change.
Keep the skill's layer descriptions and the main `README.md` Components index in sync.
