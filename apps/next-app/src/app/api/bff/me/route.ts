import { env } from "@/shared/config/env";
import { auth } from "@/shared/lib/auth";

export async function GET() {
  const session = await auth();
  const userId = session?.user?.id;

  if (!userId) {
    return Response.json({ error: "Unauthorized" }, { status: 401 });
  }

  try {
    const response = await fetch(`${env.apiBaseUrl}/users`, {
      headers: {
        "Content-Type": "application/json",
      },
    });

    if (!response.ok) {
      return Response.json(
        { error: "Failed to fetch user" },
        { status: response.status },
      );
    }

    // Find the user with matching ID from the users list
    const users = await response.json();
    const user = users.find((u: { id: number }) => u.id === Number(userId));

    if (!user) {
      return Response.json({ error: "User not found" }, { status: 404 });
    }

    return Response.json(user);
  } catch (error) {
    console.error("Error fetching user from Go API:", error);
    return Response.json({ error: "Internal server error" }, { status: 500 });
  }
}
