import { fireEvent, screen, within } from "@testing-library/react";

/**
 * Selects an option from a Material UI select rendered as a combobox.
 */
export const selectMuiOption = (selectName: string, optionName: string): void => {
  fireEvent.mouseDown(screen.getByRole("combobox", { name: selectName }));
  fireEvent.click(within(screen.getByRole("listbox")).getByRole("option", { name: optionName }));
};
