import ws from 'k6/ws';
import { check, sleep } from 'k6';

// Control-plane smoke: connect signaling WS with join token, ping, close.
// Required env: SIGNALING_WS, JOIN_TOKEN, ROOM_UUID

export const options = {
  vus: Number(__ENV.VUS || 10),
  duration: __ENV.DURATION || '30s',
};

export default function () {
  const url =
    `${__ENV.SIGNALING_WS}?room_uuid=${__ENV.ROOM_UUID}`;
  const params = {
    headers: { Authorization: `Bearer ${__ENV.JOIN_TOKEN}` },
  };
  const res = ws.connect(url, params, function (socket) {
    socket.on('open', () => {
      socket.send(JSON.stringify({ type: 'ping', room_uuid: __ENV.ROOM_UUID, payload: {}, ts: new Date().toISOString() }));
    });
    socket.on('message', (data) => {
      const msg = JSON.parse(data);
      if (msg.type === 'welcome' || msg.type === 'pong') {
        socket.close();
      }
    });
    socket.setTimeout(() => socket.close(), 5000);
  });
  check(res, { 'status is 101': (r) => r && r.status === 101 });
  sleep(1);
}
