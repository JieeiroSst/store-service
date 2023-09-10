const API_KEY = "";
const SECRET = "";

export const createNewRoom = async () => {
  const options = { 
    expiresIn: '120m', 
    algorithm: 'HS256' 
  };
   const payload = {
    apikey: API_KEY,
    permissions: [`allow_join`], // `ask_join` || `allow_mod` 
    version: 2,
    roles: ['CRAWLER'],
  };

  const token = jwt.sign(payload, SECRET, options);

  const res = await fetch(`https://api.videosdk.live/v2/rooms`, {
    method: "POST",
    headers: {
      authorization: `${token}`,
      "Content-Type": "application/json",
    },
  });

  const { roomId } = await res.json();
  return roomId;
};