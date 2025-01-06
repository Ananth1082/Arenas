"use client";

import { useEffect, useRef, useState } from "react";

export default function Page() {
  const [counterPlayer1, setCounterPlayer1] = useState(0);
  const [counterPlayer2, setCounterPlayer2] = useState(0);
  let ws: WebSocket;
  let handleClick = () => {};
  useEffect(() => {
    ws = new WebSocket("ws://localhost:8080/ws/tug-of-war");

    ws.onopen = () => {
      console.log("Connected");
      const data = localStorage.getItem("matchId");
      const id = localStorage.getItem("id");
      ws.send(
        JSON.stringify({
          userId: id,
          matchId: data,
        })
      );

      handleClick = () => {
        setCounterPlayer1(counterPlayer1 + 1);
        ws.send("PING");
      };
    };
    ws.onmessage = (event) => {
      console.log(event.data);
    };

    ws.onclose = () => {
      console.log("Disconnected");
    };
  }, []);

  return (
    <>
      <div className="flex">
        <div className="w-full h-screen flex flex-col justify-center items-center border-r-2 border-black">
          <h1>{counterPlayer1}</h1>
          <button className="border-2 border-black p-4" onClick={handleClick}>
            Player 1
          </button>
        </div>
        <div></div>
        <div className="w-full h-screen flex flex-col justify-center items-center">
          <h1>{counterPlayer2}</h1>
          <button className="border-2 border-black p-4">Player 2</button>
        </div>
      </div>
    </>
  );
}
