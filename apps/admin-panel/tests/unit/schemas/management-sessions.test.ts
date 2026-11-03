import { sessionsResponseSchema } from "@schemas/management/sessions";

const inventory = {
  success: true,
  timestamp: "2026-10-31T00:00:00Z",
  data: [
    {
      id: "42",
      uid: "1000",
      user: "alice",
      seat: "seat0",
      leader: "1234",
      class: "user",
      tty: "tty2",
      type: "wayland",
      service: "gdm",
      remote: false,
      remote_host: "",
      display: "",
      desktop: "gnome",
    },
  ],
};

describe("sessions response", () => {
  test("accepts the gateway inventory", () => {
    expect(sessionsResponseSchema.parse(inventory).data[0]?.user).toBe("alice");
  });

  test.each(["--host=evil", "", "bad/id"])("rejects invalid session ID %s", (id) => {
    expect(
      sessionsResponseSchema.safeParse({ ...inventory, data: [{ ...inventory.data[0], id }] })
        .success,
    ).toBe(false);
  });
});
