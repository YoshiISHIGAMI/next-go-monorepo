import NextAuth from "next-auth";
import GitHub from "next-auth/providers/github";
import type { components } from "@/shared/api/types.gen";
import { env } from "@/shared/config/env";

type OAuthCallbackResponse = components["schemas"]["OAuthCallbackResponse"];

export const { handlers, signIn, signOut, auth } = NextAuth({
  providers: [GitHub],
  callbacks: {
    async signIn({ user, account }) {
      if (!account || !user.email) {
        return false;
      }

      try {
        // Call Go API to create/retrieve user
        const response = await fetch(`${env.apiBaseUrl}/auth/oauth/callback`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            provider: account.provider,
            provider_account_id: account.providerAccountId,
            email: user.email,
            name: user.name ?? "",
          }),
        });

        if (!response.ok) {
          console.error(
            "Failed to sync user with Go API:",
            await response.text(),
          );
          return false;
        }

        const data: OAuthCallbackResponse = await response.json();
        // Store internal user ID in the user object for session callback
        user.id = String(data.user.id);

        return true;
      } catch (error) {
        console.error("Error calling Go API:", error);
        return false;
      }
    },
    async jwt({ token, user }) {
      // On initial sign in, user object is available
      if (user) {
        token.internalId = user.id;
      }
      return token;
    },
    async session({ session, token }) {
      // Add internal user ID to session
      if (token.internalId) {
        session.user.id = token.internalId as string;
      }
      return session;
    },
  },
});
