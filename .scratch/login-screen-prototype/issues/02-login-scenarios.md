# 02 - Login screen: scenarios

Status: ready-for-agent
Blocked by: 01

Read `../spec.md` and "Prototype" in `CODING_STANDARDS.md` first.

## What to build

Every condition the login screen can be in, reachable from the prototype bar.

- Scenarios in the prototype bar: a screen declares its scenarios in groups, the bar lists the groups of the current route, and the fake data of the screen answers according to the scenario chosen in each group. A route that declares none shows no scenario control. The choice lasts until the page is reloaded.
- The two groups of the login screen and every scenario in them, as "Scenarios" in `../spec.md` lists: Ways in and Outcome.
- The message of each scenario, in the three languages. Each says what happened and then what to do, and is shown as an Alert above the form or as an error on its field, with an icon.
- The local provider's button, last of the provider buttons, with a badge that says it is for development.

## Acceptance criteria

- [ ] The prototype bar shows the two groups on the login route and no scenario control on a route that declares none.
- [ ] Each scenario of Ways in shows exactly the ways the table lists, with no separator when only one kind of way is on.
- [ ] Wrong code: the first two attempts say how many are left, and the third disables the boxes and leaves the action that asks for a new code.
- [ ] Expired code, Limit reached, Account not registered, Provider failed and Session ended each behave as the table says.
- [ ] Changing a scenario while on the Code step leaves the screen in a state the flow allows.
- [ ] No message reports a problem by colour alone.
- [ ] No component or screen file reads a scenario by name outside the fake data of the screen.
- [ ] Every scenario is correct in light and dark, in the three languages, at desktop and 375px widths.
