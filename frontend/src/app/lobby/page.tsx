"use client";
import { useEffect, useState } from "react";

export default function page() {
  const [error, setError] = useState<string>("");
  const [msgs, setMsgs] = useState<string>("");
  const [userId, setUserId] = useState<string>("");
  function startMatchMaking() {
    const ws = new WebSocket("http://localhost:8080/ws/match-making");
    ws.onopen = (e) => {
      const details = JSON.stringify({ id: userId, name: "test" });
      ws.send(details);
    };
    ws.onmessage = (e) => {
      setMsgs(msgs + "\n" + e.data);
    };
    ws.onerror = (e) => {
      console.log(e);

      setError(JSON.stringify(e));
    };
  }
  return (
    <>
      <div className="border-2 border-black h-96 w-96 m-auto mt-10 p-4">
        <h1>Match Making</h1>
        User ID:
        <input type="text" onChange={(e) => setUserId(e.target.value)} />
        <button onClick={startMatchMaking}>Start</button>
        {error ? <h1> ERRORS: {error}</h1> : ""}
        {msgs ? <h1> MSGS: {msgs}</h1> : ""}
      </div>
    </>
  );
}
