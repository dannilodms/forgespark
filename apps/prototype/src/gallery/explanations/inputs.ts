/** The explanatory text of the Inputs route, in English only. */
export const inputs = {
  lead: "What a person types into or chooses from.",

  field: {
    about: {
      title: "Field",
      note: "A control with its label, its description and its error. Every form is a group of fields.",
    },
    withError: {
      title: "With an error",
      note: "The error says what is wrong and carries an icon, so it does not depend on colour.",
    },
    disabled: {
      title: "Every field disabled",
      note: "The field carries the disabled state, so its label fades with its control.",
    },
    edgeCases: {
      title: "Edge cases",
      note: "A long label, a long description and a long error wrap inside the width of the field.",
    },
    separator: {
      title: "Separator between two ways",
      note: "A word on the line splits two ways of doing the same thing.",
    },
  },

  label: {
    about: {
      title: "Label",
      note: "Names a control. Inside a field, use the field's own label.",
    },
    edgeCases: {
      title: "Edge cases",
      note: "A long label wraps above its control.",
    },
  },

  input: {
    about: { title: "Input", note: "One line of text typed by a person." },
    edgeCases: {
      title: "Edge cases",
      note: "A value longer than the field scrolls inside it.",
    },
  },

  inputOTP: {
    about: {
      title: "InputOTP",
      note: "A short code typed one character per box, such as the login code. Pasting the whole code fills every box.",
    },
    filled: { title: "Filled, invalid and disabled" },
  },

  textarea: {
    about: {
      title: "Textarea",
      note: "Several lines of text typed by a person. It grows with what is typed.",
    },
    edgeCases: {
      title: "Edge cases",
      note: "A long text wraps and the field grows. A word with no space breaks inside the field.",
    },
  },

  select: {
    about: {
      title: "Select",
      note: "One choice from a short list that does not fit on the screen as buttons.",
    },
  },

  checkbox: {
    about: {
      title: "Checkbox",
      note: "One option that is on or off, or several of a list. It takes effect when the form is saved.",
    },
    several: { title: "Several of a list" },
    edgeCases: {
      title: "Edge cases",
      note: "A long label wraps beside the box, which stays on the first line.",
    },
  },

  radioGroup: {
    about: {
      title: "RadioGroup",
      note: "One choice from two to five options that all fit on the screen. For a longer list, use a select.",
    },
    withDescription: { title: "With a description for each option" },
    edgeCases: {
      title: "Edge cases",
      note: "A long label wraps beside its circle. The options stay one under the other.",
    },
  },

  switch: {
    about: {
      title: "Switch",
      note: "A setting that is on or off and takes effect at once, without a save button.",
    },
    edgeCases: {
      title: "Edge cases",
      note: "A long label wraps and the switch keeps its place at the end of the first line.",
    },
  },

  inputGroup: {
    about: {
      title: "InputGroup",
      note: "A field with something attached inside its border: an icon, a unit, a shortcut or a small button. The command palette uses it for its search field.",
    },
    attached: {
      title: "What can be attached",
      note: "A text at the start or the end, a button at the end, and a row of actions under a textarea.",
    },
    states: {
      title: "States",
      note: "The group shows the focus, the error and the disabled state of the control inside it.",
    },
    edgeCases: {
      title: "Edge cases",
      note: "A long attached text keeps its width and the control takes what is left.",
    },
  },
} as const
