import { useState, useContext } from "react";
import { AuthContext } from "../contexts/AuthContext";
import { useNavigate } from "react-router-dom";

export default function Login() {
  const [email, setEmail] = useState("");
  const { login } = useContext(AuthContext);
  const [error, setError] = useState("");
  const navigate = useNavigate();

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!email) {
      setError("Email обов’язковий");
      return;
    }
    try {
      await login(email);
      navigate("/entry")
    } catch (err) {
      setError("Не вдалося увійти");
    }
  };

  return (
    <div className="app-container">
      <h2 className="page-title">Увійти</h2>
      <div className="form-card">
        <form onSubmit={handleSubmit}>
            <div>
            <label>Email:</label>
            <input
                type="email"
                value={email}
                onChange={(e) => {
                setEmail(e.target.value);
                setError("");
                }}
            />
            </div>
            {error && <div style={{ color: "red" }}>{error}</div>}
            <button type="submit" style={{ marginTop: "1rem" }}>
            Увійти
            </button>
        </form>
      </div>
    </div>
  );
}
