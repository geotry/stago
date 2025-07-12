#version 300 es

precision mediump float;
precision mediump sampler2D;
precision mediump sampler2DShadow;
precision mediump sampler2DArrayShadow;
precision mediump sampler2DArray;

in vec2 v_tex_coords;

out vec4 fragColor;

// uniform sampler2D u_screen_texture;
// uniform sampler2DShadow u_screen_texture;
uniform sampler2DArray u_screen_texture;

void main() {
  // float depth = gl_FragCoord.z / gl_FragCoord.w;
  float z = texture(u_screen_texture, vec3(v_tex_coords, 0.0f)).r;
  // float z = texture(u_screen_texture, v_tex_coords).r;
  fragColor = vec4(z, z, z, 1.0f);
  // fragColor = vec4(texture(u_screen_texture, v_tex_coords).z);
  // float depthPow = pow(depthValue.x, 10.0);
  // fragColor.xyz = vec3(depthPow);

  // float f = 10.0f; //far plane
  // float n = 2.0f; //near plane
  // float z = (2.0f * n) / (f + n - texture(u_screen_texture, v_tex_coords).x * (f - n));

  // fragColor = vec4(z, z, z, 255);
}