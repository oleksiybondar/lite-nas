import { createTheme, ThemeProvider } from "@mui/material/styles";
import Table from "@mui/material/Table";
import TableBody from "@mui/material/TableBody";
import { SensorChartStopRow } from "@pages/SensorChartEditor/index";
import type { SensorChart } from "@schemas/monitoring/sensor";
import { fireEvent, screen } from "@testing-library/react";
import { renderWithLanguage } from "../test-utils/language";

/** Renders one chart stop with a known active theme for color choice checks. */
const renderStopRow = (
  onChange: (index: number, patch: Partial<SensorChart["stops"][number]>) => void,
) => {
  const theme = createTheme({ palette: { primary: { main: "#123456" } } });
  return renderWithLanguage(
    <ThemeProvider theme={theme}>
      <Table>
        <TableBody>
          <SensorChartStopRow
            index={0}
            onChange={onChange}
            onRemove={() => undefined}
            stop={{ value: 5, color: "#ff0000", label: "", showBar: false, showLabel: false }}
          />
        </TableBody>
      </Table>
    </ThemeProvider>,
  );
};

test("selects a color from the native picker", () => {
  const onChange = vi.fn();
  renderStopRow(onChange);
  fireEvent.change(screen.getByLabelText("Color picker"), { target: { value: "#abcdef" } });
  expect(onChange).toHaveBeenCalledWith(0, { color: "#abcdef" });
});

test("searches and selects the primary color from the active theme", () => {
  const onChange = vi.fn();
  renderStopRow(onChange);
  fireEvent.change(screen.getByRole("combobox", { name: "Color" }), {
    target: { value: "primary" },
  });
  fireEvent.click(screen.getByRole("option", { name: /Theme color: Primary/ }));
  expect(onChange).toHaveBeenCalledWith(0, { color: "#123456" });
});

test("searches and selects a common color preset", () => {
  const onChange = vi.fn();
  renderStopRow(onChange);
  fireEvent.change(screen.getByRole("combobox", { name: "Color" }), {
    target: { value: "orange" },
  });
  fireEvent.click(screen.getByRole("option", { name: /Orange/ }));
  expect(onChange).toHaveBeenCalledWith(0, { color: "#ff9800" });
});

test("accepts a custom six-digit hex color", () => {
  const onChange = vi.fn();
  renderStopRow(onChange);
  fireEvent.change(screen.getByRole("combobox", { name: "Color" }), {
    target: { value: "#abcdef" },
  });
  expect(onChange).toHaveBeenCalledWith(0, { color: "#abcdef" });
});

test("does not save a search term as a color", () => {
  const onChange = vi.fn();
  renderStopRow(onChange);
  fireEvent.change(screen.getByRole("combobox", { name: "Color" }), {
    target: { value: "orange" },
  });
  expect(onChange).not.toHaveBeenCalled();
});
