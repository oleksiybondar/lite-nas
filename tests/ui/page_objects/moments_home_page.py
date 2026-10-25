"""HyperionTF page model for the authenticated Moments home."""

from hyperiontf import By, WebPage, element
from hyperiontf.ui.element import Element


class MomentsHomePage(WebPage):  # type: ignore[misc]
    """Authenticated Moments home that contains the demo gallery region."""

    @element  # type: ignore[untyped-decorator]
    def gallery(self) -> Element:
        """Demo gallery region shown after Moments login succeeds."""
        return By.css("[data-testid='moments-gallery']")
