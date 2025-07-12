#version 300 es

void main() {
  // gl_FragDepth = gl_FragCoord.z;

  // See: https://github.com/eugene/webgl-vsm-shadows/blob/master/index.html#L92
  // float vDepth = gl_FragCoord.z;
  // float depth2 = pow(vDepth, 2.0);

  // // // approximate the spatial average of vDepth^2
  // float dx = dFdx(vDepth);
  // float dy = dFdy(vDepth);
  // float depth2Avg = depth2 + 0.50 * (dx*dx + dy*dy);

  // // depth saved in red channel while average depth^2 is
  // // stored in the green channel
  // fragColor = vec4(vDepth, depth2Avg, 0., 1.);
}