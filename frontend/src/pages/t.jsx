import { createContext, useState, useEffect } from "react";
import api from "../api/axios";

export const AuthContext = createContext({});

export function AuthProvider({ children }) {
  // Спочатку беремо з localStorage або "".
  const [token, setToken] = useState("");

  // Після першого монтування читаємо токен із localStorage і кладемо в state
  useEffect(() => {
    const savedToken = localStorage.getItem("token");
    if (savedToken) {
      setToken(savedToken);
    }
  }, []);

  // робимо запит, отримуємо токен, зберігаємо в state + localStorage
  const login = async (email) => {
    const resp = await api.post("/auth/login", { email });
    const newToken = resp.data.token;
    localStorage.setItem("token", newToken);
    setToken(newToken);
  };

  // очищаємо токен із state і з localStorage
  const logout = () => {
    localStorage.removeItem("token");
    setToken("");
  };

  return (
    <AuthContext.Provider value={{ token, login, logout }}>
      {children}
    </AuthContext.Provider>
  );
}
