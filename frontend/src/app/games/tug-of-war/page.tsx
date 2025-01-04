"use client";

import { useEffect, useState } from "react";

export default function Page() {
  const [counterPlayer1, setCounterPlayer1] = useState(0);
  const [counterPlayer2, setCounterPlayer2] = useState(0);
  useEffect(() => {});
  return (
    <>
      <div className="flex">
        <div className="w-full h-screen flex flex-col justify-center items-center border-r-2 border-black">
          <h1>{counterPlayer1}</h1>
          <button
            className="border-2 border-black p-4"
            onClick={() => {
              setCounterPlayer1(counterPlayer1 + 1);
            }}
          >
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
