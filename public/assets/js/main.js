var Module = { canvas: document.getElementById('canvas'), arguments: [] };

const my_comms = () => {
    const smaller_canvas = document.getElementById('smaller-canvas');
    const smaller_ctx = smaller_canvas.getContext('2d', { willReadFrequently: true });
    console.log('my_comms start');
    const can = Module.canvas;
    const ctx = can.getContext('webgl');
    // const ctx = canvas.getContext("2d", { willReadFrequently: true });

    const W = can.width;
    const H = can.height;
    console.log('can', can, 'ctx', ctx);
    const DRAW_W = ctx.drawingBufferWidth;
    const DRAW_H = ctx.drawingBufferHeight;
    // const DRAW_W = 128;
    // const DRAW_H = DRAW_W;
    console.log('DRAW_W', DRAW_W, 'DRAW_H', DRAW_H);
    const RGB = ctx.RGB;
    const RGBA = ctx.RGBA;
    const size = W * H * 3;
    const new_size = DRAW_W * DRAW_H * 4;
    const UNSIGNED_BYTE = ctx.UNSIGNED_BYTE;
    console.log('W', W, 'H', H, 'size', size, 'new_size', new_size);
    console.log('RGB', RGB, 'RGBA', RGBA, 'UNSIGNED_BYTE', UNSIGNED_BYTE);
    // const pixels = new Uint8Array(new_size);
    // ctx.readPixels(0, 0, DRAW_W, DRAW_H, RGBA, UNSIGNED_BYTE, pixels);
    // ctx.readPixels(0, 0, DRAW_H, DRAW_H, RGBA, UNSIGNED_BYTE, pixels);
    // console.log('pixels', pixels);

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
    // const smaller_img_data = smaller_ctx.createImageData(DRAW_W, DRAW_H);
    const setup = () => {
        console.log("setup start");
        let last_t = 0;
        const step = (t) => {
            requestAnimationFrame(step);
            if (t - last_t < 1) return;
            last_t = t;
            // ctx.readPixels(0, 0, DRAW_W, DRAW_H, RGBA, UNSIGNED_BYTE, pixels);
            // ctx.readPixels(0, 0, DRAW_H, DRAW_H, RGBA, UNSIGNED_BYTE, pixels);
            // smaller_ctx.drawImage(can, 0, 0, DRAW_W, DRAW_H, 0, 0, 128, 128);
            smaller_ctx.drawImage(can, 0, 0, DRAW_H, DRAW_H, 0, 0, 128, 128);
            const smaller_img_data = smaller_ctx.getImageData(0, 0, 128, 128);
            // console.log('smaller_img_data:', smaller_img_data);
            ws.send(smaller_img_data.data);
            // console.log('send data');
        };
        requestAnimationFrame(step);
        console.log("setup end");
    };
    ws.addEventListener('error', (e) => console.error('websocket error:', e));
    ws.addEventListener('open', (e) => {
        console.log('websocket open:', e);
        setup();
    });
    ws.addEventListener('close', (e) => console.log('websocket close:', e));
    ws.addEventListener('message', handler_msg);

    console.log('my_comms end');
};

const main = () => {
    const gameFrame = document.getElementById('game-frame');
    gameFrame.addEventListener('click', () => {
        const scriptTag = document.createElement('script'); // create a script tag
        const firstScriptTag = document.getElementsByTagName('script')[0]; // find the first script tag in the document
        scriptTag.src = '/assets/js/tic80.js'; // set the source of the script to your script
        firstScriptTag.parentNode.insertBefore(scriptTag, firstScriptTag); // append the script to the DOM
        gameFrame.remove();
        setTimeout(my_comms, 5000);
    });
};

main();
