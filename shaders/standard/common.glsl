// uniform int u_something;

// uniform Camera {
//   mat4 view;
//   mat4 projection;
// };

float linstep(float low, float high, float v) {
  return clamp((v - low) / (high - low), 0.0, 1.0);
}