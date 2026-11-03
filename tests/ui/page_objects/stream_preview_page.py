"""HyperionTF page model for a browser-rendered Stream Service preview."""

from hyperiontf import By, WebPage, element
from hyperiontf.ui.element import Element


class StreamPreviewPage(WebPage):  # type: ignore[misc]
    """Browser image document produced by an authorized HEIC JPEG preview."""

    @element  # type: ignore[untyped-decorator]
    def image(self) -> Element:
        """Native browser image element containing the decoded JPEG response."""
        return By.css("img")

    def image_rendered(self, minimum_size: int) -> bool:
        """Check that the native image document displays at the expected size."""
        self.image.wait_until_visible()
        size = self.image.get_size()
        return bool(size["width"] >= minimum_size and size["height"] >= minimum_size)
