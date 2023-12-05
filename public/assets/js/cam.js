const W = 128;
const H = W;

const get_cam_stream = (constraints = true) => navigator.mediaDevices.getUserMedia({ video: constraints, audio: false });

const main = async () => {
    console.log('main start');
    // websockets
    const ws = new WebSocket(`ws://${window.location.host}/ws`);
    ws.binaryType = "arraybuffer";
    const handler_msg = (e) => {
        console.log('websocket message:', e);
        const is_binary = e.data instanceof ArrayBuffer;
        console.log('is_binary', is_binary);
        if (is_binary) {
            const dec = new TextDecoder('utf-8');
            const s = dec.decode(e.data);
            console.log('decode binary message as utf-8 string:', s);
            ws.send('hello from websocket client');
        }
    };
    ws.addEventListener('error', (e) => console.error('websocket error:', e));
    ws.addEventListener('message', handler_msg);
    const wait_websocket = new Promise((resolve, reject) => {
        ws.addEventListener('close', (e) => { console.log('websocket close:', e); reject(); });
        ws.addEventListener('open', (e) => { console.log('websocket open:', e); resolve(); });
    });

    // canvas
    const can = document.getElementById('output-canvas');
    can.width = W;
    can.height = W;
    const ctx = can.getContext('2d', { willReadFrequently: true });
    ctx.fillStyle = 'black';
    ctx.fillRect(0, 0, W, H);

    const cam_stream = await get_cam_stream({ width: 480, height: 480 });
    const video = document.getElementById('output-video');
    video.srcObject = cam_stream;
    const wait_playing = new Promise((resolve) => {
        video.addEventListener('playing', () => {
            console.log('started playing');
            resolve();
        });
    });
    video.play();
    await wait_playing;
    await wait_websocket;

    const VIDEO_W = video.videoWidth;
    const VIDEO_H = video.videoHeight;
    console.log('VIDEO_W', VIDEO_W, 'VIDEO_H', VIDEO_H);

    let last_t = 0;
    const step = (t) => {
        requestAnimationFrame(step);
        if (t - last_t < 1) return;
        last_t = t;
        ctx.drawImage(video, 0, 0, VIDEO_W, VIDEO_H, 0, 0, W, H);
        const img_data = ctx.getImageData(0, 0, W, H);
        ws.send(img_data.data);
    };
    requestAnimationFrame(step);
    console.log('main end');
};

main().catch(console.error);
/*
  .then((stream) => {
    video.srcObject = stream;
    video.play();
  })
  .catch((err) => {
    console.error(`An error occurred: ${err}`);
  });
*/
