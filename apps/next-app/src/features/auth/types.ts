export type User = {
  id: number;
  email: string;
  createdAt: string;
};

export type AuthSession = {
  user: User;
  accessToken: string;
};
