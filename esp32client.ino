#include <WiFi.h>
#include <FastLED.h>
#include <ESP32-HUB75-MatrixPanel-I2S-DMA.h>

#define PIN_E 32
#define PANELS_NUMBER 4
#define PANEL_WIDTH 64
#define PANEL_HEIGHT 64
#define NUM_LEDS (PANEL_WIDTH * PANEL_HEIGHT)
#define BYTES_NUM_LEDS (NUM_LEDS * 3)
#define TIMEOUT 3000

// For debug messages over Serial port
String helper_string = "";

const char* ssid      = "xxxx";
const char* password  = "xxxx";
const uint16_t port = 52275; // port TCP server
const char * host = "192.168.0.3"; // ip or dns
int curr_time = -1;
int prev_time = curr_time; // not connected

WiFiClient client;

CRGB leds[NUM_LEDS];
uint8_t *leds_int = (uint8_t *)leds;
MatrixPanel_I2S_DMA *dma_display = nullptr;
int current_panel = 0;
int n_bytes_so_far = -1;

void setup() {
  Serial.begin(115200);
  WiFi.begin(ssid, password);
  while (WiFi.status() != WL_CONNECTED) {
    delay(500);
    Serial.println("Connecting to WiFi..");
  }
  Serial.println("");
  Serial.println("WiFi connected");
  Serial.println("IP address: ");
  Serial.println(WiFi.localIP());

  // https://github.com/mrfaptastic/ESP32-HUB75-MatrixPanel-DMA/blob/722358ad2d990d0aa36600da95e4b36b740ff5f7/src/ESP32-HUB75-MatrixPanel-I2S-DMA.h#L232
  HUB75_I2S_CFG mxconfig;
  mxconfig.mx_width = PANEL_WIDTH;
  mxconfig.mx_height = PANEL_HEIGHT;
  mxconfig.chain_length = PANELS_NUMBER;
  mxconfig.gpio.e = PIN_E;
  dma_display = new MatrixPanel_I2S_DMA(mxconfig);
  dma_display->setBrightness8(192);
  if (!(dma_display->begin())) {
    Serial.println("****** !KABOOM! I2S memory allocation failed ***********");
    return;
  }
  Serial.println("Fill screen: Neutral White");
  dma_display->fillScreenRGB888(64, 64, 64);
  delay(1000);
  Serial.println("Fill screen: black");
  dma_display->fillScreenRGB888(0, 0, 0);
  delay(1000);
  Serial.println("Starting the loop.....");
}

void loop()
{
  curr_time = millis();
  if (prev_time < 0)
  {
    Serial.print("Connecting to ");
    Serial.println(host);
    if (!client.connect(host, port))
    {
      Serial.println("Connection failed.");
      Serial.println("Waiting 5 seconds before retrying...");
      delay(5000);
      return;
    }
    prev_time = curr_time;
  }
  if (n_bytes_so_far < 0)
  {
    n_bytes_so_far = 0;
    // Serial.println(helper_string + "requesting current_panel: " + current_panel);
    client.print(current_panel);
  }
  if (client.available() <= 0)
  {
    // Serial.println("no bytes available from client.available()");
    if (curr_time - prev_time > TIMEOUT)
    {
      Serial.println("Closing connection.");
      prev_time = -1;
      n_bytes_so_far = -1;
      client.stop();
    }
    return;
  }
  prev_time = curr_time;
  // Serial.println("loop - read the frame data");
  int n_bytes = client.read(leds_int + n_bytes_so_far, BYTES_NUM_LEDS - n_bytes_so_far);
  n_bytes_so_far += n_bytes;
  // Serial.println(helper_string + "after reading n_bytes: " + n_bytes + " n_bytes_so_far: " + n_bytes_so_far);
  if (n_bytes_so_far < BYTES_NUM_LEDS)
  {
    return;
  }
  n_bytes_so_far = -1;
  // Serial.println("before drawing");
  int panel_offset = current_panel * PANEL_WIDTH;
  current_panel = (current_panel + 1) % 4;
  for (int y = 0, pixel_idx = 0; y < PANEL_HEIGHT; y++)
  {
    for (int x = 0; x < PANEL_WIDTH; x++)
    {
      CRGB currentColor = leds[pixel_idx++];
      dma_display->drawPixelRGB888(panel_offset + x, y, currentColor.r, currentColor.g, currentColor.b);
    }
  }
  // Serial.println("after drawing");
}