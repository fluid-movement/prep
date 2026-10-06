## D1: Tag syntax
date: 2026-10-06

Tags are lowercase letters, digits and . _ - /, starting with a letter or digit, so they are easy to type in filters and never need quoting. Input is lowercased; anything else is rejected before writing and reported by prep check (I026). Repeated --tag filters OR together like other repeated flags.

## D2: Tags stay editable on resolved issues
date: 2026-10-06

Tags have no lifecycle meaning and are most useful for navigating past work, so prep edit --tag alone works on done and dropped issues; every other edit still requires an unresolved issue (decision D1 of 20261005-184730 stands for the record itself).
