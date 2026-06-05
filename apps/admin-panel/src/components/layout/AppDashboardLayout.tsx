import { AppChromeLayout } from "@components/layout/AppChromeLayout";
import { AppAlertsControl } from "@components/navigation/AppAlertsControl";
import { AppFooter } from "@components/navigation/AppFooter";
import { AppSidebar } from "@components/navigation/AppSidebar";
import { AppSidebarDrawer, AppSidebarDrawerButton } from "@components/navigation/AppSidebarDrawer";
import { AppSidebarFlyout } from "@components/navigation/AppSidebarFlyout";
import { AppSidebarModeToggle } from "@components/navigation/AppSidebarModeToggle";
import { AppTopBar } from "@components/navigation/AppTopBar";
import { AppUserMenu } from "@components/navigation/AppUserMenu";
import Box from "@mui/material/Box";
import Container from "@mui/material/Container";
import {
  type AppNavigationItem,
  resolveNavigationItems,
  resolveSelectedNavigationPath,
} from "@routes/navigation";
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
  const navigationItems = resolveNavigationItems(pathname);
  const selectedPath = resolveSelectedNavigationPath(pathname, navigationItems);
  const [isMobileSidebarOpen, setIsMobileSidebarOpen] = useState(false);
  const [isSidebarCollapsed, setIsSidebarCollapsed] = useState(false);

  return (
    <AppChromeLayout
      footer={<AppFooter />}
      header={renderDashboardHeader({
        isSidebarCollapsed,
        onOpenMobileSidebar: () => {
          setIsMobileSidebarOpen(true);
        },
        onToggleSidebarMode: () => {
          setIsSidebarCollapsed((currentValue) => !currentValue);
        },
      })}
      main={renderDashboardMain({
        isMobileSidebarOpen,
        isSidebarCollapsed,
        navigationItems,
        onCloseMobileSidebar: () => {
          setIsMobileSidebarOpen(false);
        },
        selectedPath,
      })}
    />
  );
};

/**
 * State and commands required to render dashboard header controls.
 */
type DashboardHeaderRenderOptions = {
  isSidebarCollapsed: boolean;
  onOpenMobileSidebar: () => void;
  onToggleSidebarMode: () => void;
};

/**
 * Builds the dashboard top bar with sidebar controls and authenticated user menu.
 */
const renderDashboardHeader = ({
  isSidebarCollapsed,
  onOpenMobileSidebar,
  onToggleSidebarMode,
}: DashboardHeaderRenderOptions): ReactElement => {
  return (
    <AppTopBar
      leadingAction={
        <Box alignItems="center" display="flex" gap={1}>
          <AppSidebarDrawerButton onOpen={onOpenMobileSidebar} />
          <AppSidebarModeToggle isCollapsed={isSidebarCollapsed} onToggle={onToggleSidebarMode} />
        </Box>
      }
      trailingAction={
        <Box alignItems="center" display="flex" gap={1}>
          <AppAlertsControl />
          <AppUserMenu />
        </Box>
      }
    />
  );
};

/**
 * State and commands required to render dashboard navigation and content.
 */
type DashboardMainRenderOptions = {
  isMobileSidebarOpen: boolean;
  isSidebarCollapsed: boolean;
  navigationItems: AppNavigationItem[];
  onCloseMobileSidebar: () => void;
  selectedPath: string | null;
};

/**
 * Builds the dashboard sidebar variants around the routed content outlet.
 */
const renderDashboardMain = ({
  isMobileSidebarOpen,
  isSidebarCollapsed,
  navigationItems,
  onCloseMobileSidebar,
  selectedPath,
}: DashboardMainRenderOptions): ReactElement => {
  return (
    <Box data-testid="dashboard-layout" display="flex" minHeight="calc(100vh - 113px)">
      <AppSidebarDrawer
        items={navigationItems}
        onClose={onCloseMobileSidebar}
        open={isMobileSidebarOpen}
        selectedPath={selectedPath}
      />
      <AppSidebarFlyout
        display={{ md: isSidebarCollapsed ? "block" : "none", xs: "none" }}
        items={navigationItems}
        selectedPath={selectedPath}
      />
      <AppSidebar
        display={{ md: isSidebarCollapsed ? "none" : "block", xs: "none" }}
        items={navigationItems}
        selectedPath={selectedPath}
      />
      <Container
        component="section"
        data-testid="dashboard-content"
        maxWidth={false}
        sx={{ py: 4 }}
      >
        <Outlet />
      </Container>
    </Box>
  );
};
