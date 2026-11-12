// biome-ignore-all lint/complexity/noExcessiveCognitiveComplexity: responsive layout branches are the shared tab component's contract.
import Box from "@mui/material/Box";
import { useTheme } from "@mui/material/styles";
import Tabs from "@mui/material/Tabs";
import useMediaQuery from "@mui/material/useMediaQuery";
import type { ReactElement, ReactNode } from "react";
import { z } from "zod";

/** Styled tab navigation shared by routed and local preferences screens. */
export const PreferencesTabs = ({
  label,
  value,
  children,
  layout = "inline",
  onChange,
  testId,
}: {
  label: string;
  value: string;
  children: ReactNode;
  layout?: "inline" | "sidebar";
  onChange?: (value: string) => void;
  testId?: string;
}): ReactElement => {
  const theme = useTheme();
  const wide = useMediaQuery(theme.breakpoints.up("sm"));
  const vertical = layout === "sidebar" && wide;
  const horizontalVariant = layout === "inline" ? "fullWidth" : "scrollable";

  return (
    <Box
      component="nav"
      data-testid={testId}
      sx={{
        borderBottom: vertical ? 0 : 1,
        borderRight: vertical ? 1 : 0,
        borderColor: "divider",
        flexShrink: 0,
        width: vertical ? 220 : "100%",
      }}
    >
      <Tabs
        aria-label={label}
        orientation={vertical ? "vertical" : "horizontal"}
        value={value}
        variant={vertical ? "standard" : horizontalVariant}
        onChange={(_, next: unknown) => {
          const parsed = z.string().safeParse(next);
          if (parsed.success) onChange?.(parsed.data);
        }}
        sx={{
          alignItems: "stretch",
          "& .MuiTab-root": {
            minHeight: 48,
            px: 2,
            textTransform: "none",
            ...(vertical && { alignItems: "center", justifyContent: "flex-start" }),
          },
        }}
      >
        {children}
      </Tabs>
    </Box>
  );
};
