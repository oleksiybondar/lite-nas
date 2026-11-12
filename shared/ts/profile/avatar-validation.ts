import { z } from "zod";

/** Browser upload contract for fixed PNG/JPEG avatars. */
export const avatarFileSchema = z
  .custom<File>((value) => typeof File !== "undefined" && value instanceof File)
  .refine(
    (file) =>
      file.type === "image/png" || file.type === "image/jpeg" || /\.(png|jpe?g)$/i.test(file.name),
    "Choose a PNG or JPEG image.",
  )
  .refine(
    (file) => file.size > 0 && file.size <= 5 * 1024 * 1024,
    "Image must be 5 MB or smaller.",
  );
