import { AppChromeLayout } from "@components/layout/AppChromeLayout";
import { AppFooter } from "@components/navigation/AppFooter";
import { AppSidebar } from "@components/navigation/AppSidebar";
import { AppSidebarDrawer, AppSidebarDrawerButton } from "@components/navigation/AppSidebarDrawer";
import { AppSidebarFlyout } from "@components/navigation/AppSidebarFlyout";
import { AppSidebarModeToggle } from "@components/navigation/AppSidebarModeToggle";
import { AppTopBar } from "@components/navigation/AppTopBar";
import Box from "@mui/material/Box";
import Container from "@mui/material/Container";
import { appNavigationItems, resolveSelectedNavigationPath } from "@routes/navigation";
import type { ReactElement } from "react";
import { useState } from "react";
import { Outlet, useLocation } from "react-router-dom";

/**
 * Protected dashboard layout for authenticated admin-panel routes.
 *
 * This is the current Material-only equivalent of the intended Toolpad Core
 * dashboard frame. It keeps header/footer in shared chrome slots and owns only
 * the protected sidebar + routed main content.
 */
export const AppDashboardLayout = (): ReactElement => {
  const { pathname } = useLocation();
  const selectedPath = resolveSelectedNavigationPath(pathname);
  const [isMobileSidebarOpen, setIsMobileSidebarOpen] = useState(false);
  const [isSidebarCollapsed, setIsSidebarCollapsed] = useState(false);

  return (
    <AppChromeLayout
      footer={<AppFooter />}
      header={
        <AppTopBar
          leadingAction={
            <Box alignItems="center" display="flex" gap={1}>
              <AppSidebarDrawerButton
                onOpen={() => {
                  setIsMobileSidebarOpen(true);
                }}
              />
              <AppSidebarModeToggle
                isCollapsed={isSidebarCollapsed}
                onToggle={() => {
                  setIsSidebarCollapsed((currentValue) => !currentValue);
                }}
              />
            </Box>
          }
        />
      }
      main={
        <Box display="flex" minHeight="calc(100vh - 113px)">
          <AppSidebarDrawer
            items={appNavigationItems}
            onClose={() => {
              setIsMobileSidebarOpen(false);
            }}
            open={isMobileSidebarOpen}
            selectedPath={selectedPath}
          />
          <AppSidebarFlyout
            display={{ md: isSidebarCollapsed ? "block" : "none", xs: "none" }}
            items={appNavigationItems}
            selectedPath={selectedPath}
          />
          <AppSidebar
            display={{ md: isSidebarCollapsed ? "none" : "block", xs: "none" }}
            items={appNavigationItems}
            selectedPath={selectedPath}
          />
          <Container component="section" maxWidth={false} sx={{ py: 4 }}>
            <Outlet />
          </Container>
        </Box>
      }
    />
  );
};
