import { UserMenu } from "@lite-nas/profile";
import { fireEvent, render, screen } from "@testing-library/react";

describe("shared user menu", () => {
  test.each([
    { label: "User profile", action: "profile" },
    { label: "Application settings", action: "settings" },
    { label: "Logout", action: "logout" },
  ])("runs the hosting app's $action action", ({ label, action }) => {
    const calls: string[] = [];
    render(
      <UserMenu
        login="alex"
        onProfile={() => calls.push("profile")}
        onApplicationSettings={() => calls.push("settings")}
        onLogout={() => calls.push("logout")}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: "User menu" }));
    fireEvent.click(screen.getByRole("menuitem", { name: label }));

    expect(calls).toEqual([action]);
  });
});
