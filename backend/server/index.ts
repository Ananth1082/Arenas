import express from "express";

const app = express();
const port = 8080;
app.get("/ping", (_, res) => {
  res.status(200).send("hello");
});
app.listen(port, () => console.log(`Server running at port ${port}`));
