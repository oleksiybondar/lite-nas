import { z } from "zod";

/** Validates the USBGuard inventory envelope before parsing its textual records. */
export const usbDevicesResponseSchema = z.object({
  success: z.literal(true),
  timestamp: z.string(),
  data: z.string(),
});

/** Parses one USBGuard list-devices record into fields suitable for the inventory UI. */
export const usbDeviceSchema = z.object({
  id: z.string().regex(/^[0-9]{1,10}$/),
  product: z.string(),
  hash: z.string(),
  name: z.string(),
  state: z.enum(["allow", "block", "reject", "unknown"]),
  attributes: z.string(),
});

/** Validates a USBGuard policy decision before the provider sends it. */
export const usbDeviceActionSchema = z.object({
  permanent: z.boolean(),
});

export type USBDevice = z.infer<typeof usbDeviceSchema>;

/** Converts USBGuard's line-oriented inventory into validated device records. */
export const parseUSBDevices = (value: unknown): USBDevice[] => {
  const { data } = usbDevicesResponseSchema.parse(value);
  if (data.trim() === "") return [];
  return data.split(/\r?\n/).map((line) => {
    const match = /^(\d+)\s*:\s*(allow|block|reject)\s+([\s\S]+)$/.exec(line.trim());
    if (!match) throw new Error("Malformed USBGuard device record");
    const [, id = "", rawState = "unknown", attributes = ""] = match;
    const name = /(?:^|\s)name\s+"((?:[^"\\]|\\.)*)"/.exec(attributes)?.[1] ?? "";
    const product = /(?:^|\s)id\s+([0-9a-fA-F]{4}:[0-9a-fA-F]{4})/.exec(attributes)?.[1] ?? "";
    const hash = /(?:^|\s)hash\s+"?([0-9a-fA-F]+)"?/.exec(attributes)?.[1] ?? "";
    return usbDeviceSchema.parse({
      id,
      product,
      hash,
      name: name.replace(/\\"/g, '"'),
      state: rawState,
      attributes,
    });
  });
};
