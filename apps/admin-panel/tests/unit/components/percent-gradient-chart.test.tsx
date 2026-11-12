import { PercentGradientChart } from "@components/monitoring/PercentGradientChart";
import { render } from "@testing-library/react";

/**
 * Returns the percent chart fill gradient and fails the test when it is not rendered.
 */
const mustQueryPercentGradientFill = (container: HTMLElement): SVGLinearGradientElement => {
  const gradient = container.querySelector<SVGLinearGradientElement>(
    "#percent-gradient-chart-fill",
  );

  expect(gradient).not.toBeNull();

  return gradient as SVGLinearGradientElement;
};

describe("PercentGradientChart", () => {
  test("anchors the fill gradient to the fixed percent domain", () => {
    const { container } = render(
      <PercentGradientChart capacity={2} stamps={["t1", "t2"]} values={[18, 20]} />,
    );

    const gradient = mustQueryPercentGradientFill(container);

    expect(gradient).toHaveAttribute("gradientUnits", "userSpaceOnUse");
    expect(gradient).toHaveAttribute("y1", "186");
    expect(gradient).toHaveAttribute("y2", "12");
  });

  test("keeps threshold color stops on percent offsets", () => {
    const { container } = render(
      <PercentGradientChart capacity={2} stamps={["t1", "t2"]} values={[18, 20]} />,
    );

    const gradient = mustQueryPercentGradientFill(container);
    const stops = Array.from(gradient.querySelectorAll("stop")).map((stop) => {
      return {
        color: stop.getAttribute("stop-color"),
        offset: stop.getAttribute("offset"),
      };
    });

    expect(stops).toEqual([
      { color: "#2563eb", offset: "0%" },
      { color: "#2563eb", offset: "25%" },
      { color: "#16a34a", offset: "50%" },
      { color: "#ea580c", offset: "75%" },
      { color: "#dc2626", offset: "100%" },
    ]);
  });
});
