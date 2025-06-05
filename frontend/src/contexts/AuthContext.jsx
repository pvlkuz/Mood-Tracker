import { createContext, useState, useEffect } from "react";
import api from "../api/axios";
import { useNavigate } from "react-router-dom";

export const AuthContext = createContext();

export function AuthProvider({ children }) {
  const navigate = useNavigate();
  const [token, setToken] = useState(localStorage.getItem("token") || "");

  // Якщо в localStorage є токен, ми вже вважаємося залогіненими
  useEffect(() => {
    if (token) {
      localStorage.setItem("token", token);
    } else {
      localStorage.removeItem("token");
    }
  }, [token]);

  // робить "POST /auth/login" і зберігає токен в localStorage
  const login = async (email) => {
    const resp = await api.post("/auth/login", { email });
    setToken(resp.data.token);
    navigate("/entry");
  };

  const logout = () => {
    localStorage.removeItem("token");
    setToken("");
    navigate("/login");
  };

  return (
    <AuthContext.Provider value={{ token, login, logout }}>
      {children}
    </AuthContext.Provider>
  );
}
