import { Link, useNavigate } from "react-router-dom";
import { useContext } from "react";
import { AuthContext } from "../contexts/AuthContext";

export default function Navbar() {
  const { token, logout } = useContext(AuthContext);
  const navigate = useNavigate();

  const handleLogout = () => {
    logout();           // очищаємо token і localStorage
    navigate("/login"); // переадресація на /login
  };

  return (
    <nav className="navbar">
      {token ? (
        <>
          <Link to="/entry">Сьогоднішній настрій</Link>
          <Link to="/history">Історія</Link>
          <Link to="/chart">Графік</Link>
          <Link to="/telegram">Підключити Telegram</Link>
          <Link className="logout-button" onClick={handleLogout}>Вийти</Link>
        </>
      ) : (
        <Link to="/login">Логін</Link>
      )}
    </nav>
  );
}
