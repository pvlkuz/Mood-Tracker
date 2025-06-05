import { useState } from "react";


//IconPicker відображає набір іконок, викликає onSelect(icon) коли користувач обрав іконку.
export default function IconPicker({ onSelect }) {
  const icons = ["😃", "😊", "😐", "😞", "😢", "😡"];
  const [selected, setSelected] = useState("");

  const handleClick = (icon) => {
    setSelected(icon);
    onSelect(icon);
  };

  return (
    <div style={{ display: "flex", gap: "1rem", marginTop: "1rem" }}>
      {icons.map((icon) => (
        <div
          key={icon}
          onClick={() => handleClick(icon)}
          style={{
            fontSize: "2rem",
            padding: "0.5rem",
            cursor: "pointer",
            border: selected === icon ? "2px solid blue" : "1px solid #ccc",
            borderRadius: "8px",
          }}
        >
          {icon}
        </div>
      ))}
    </div>
  );
}
