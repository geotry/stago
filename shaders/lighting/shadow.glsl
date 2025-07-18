float shadow_filter(sampler2DShadow shadowMap, vec3 uv_shadowMap, vec2 shadowMapSize) {
    // return texture(u_shadow_map, uv_shadowMap);
  float result = 0.0;

  for(int x = -3; x <= 3; x++) {
    for(int y = -3; y <= 3; y++) {
      float x_l = (uv_shadowMap.x - float(x) / float(shadowMapSize.x));
      float y_l = (uv_shadowMap.y - float(y) / float(shadowMapSize.y));
      vec3 lookup = vec3(x_l, y_l, uv_shadowMap.z);
      result += texture(shadowMap, lookup); //get(x,y);
    }
  }

  return result / 49.0;
}

float shadow_filter(sampler2DArrayShadow shadowMap, int index, vec3 uv_shadowMap, vec2 shadowMapSize) {
    // return texture(shadowMap, vec4(uv_shadowMap, float(index)));
  float result = 0.0;

  for(int x = -3; x <= 3; x++) {
    for(int y = -3; y <= 3; y++) {
      float x_l = (uv_shadowMap.x - float(x) / float(shadowMapSize.x));
      float y_l = (uv_shadowMap.y - float(y) / float(shadowMapSize.y));
      vec3 lookup = vec3(x_l, y_l, uv_shadowMap.z);
      result += texture(shadowMap, vec4(lookup, float(index)));
    }
  }

  return result / 49.0;
}

float ShadowCalculation(in sampler2DShadow shadowMap, vec4 fragPosLightSpace, float bias) {
  vec3 projCoords = fragPosLightSpace.xyz / fragPosLightSpace.w;
  projCoords = projCoords * 0.5 + 0.5;
    // vec3 projCoords = fragPosLightSpace.xyz;

    // float closestDepth = texture(shadowMap, projCoords); 
    // // get depth of current fragment from light's perspective
    // float currentDepth = projCoords.z;
    // // check whether current frag pos is in shadow
    // float shadow = currentDepth > closestDepth ? 0.0f : 1.0f;
    // return shadow;

  vec2 shadowMapSize = vec2(textureSize(shadowMap, 0));
    // return textureProj(shadowMap, fragPosLightSpace);
    // return textureProjOffset(shadowMap, vec4(projCoords, fragPosLightSpace.w), ivec2(.5f));
    // return textureProj(shadowMap, vec4(projCoords, fragPosLightSpace.w), 0.5f);

    // return texture(shadowMap, projCoords);

  float shadowCoeff;

    // BEGIN PCF
  float sum = 0.0;
  vec2 duv;

  for(float pcf_x = -1.5; pcf_x <= 1.5; pcf_x += 1.) {
    for(float pcf_y = -1.5; pcf_y <= 1.5; pcf_y += 1.) {
      duv = vec2(pcf_x / float(shadowMapSize.x), pcf_y / float(shadowMapSize.y));
      sum += shadow_filter(shadowMap, vec3(projCoords.xy, projCoords.z) + vec3(duv, 0.0), shadowMapSize);
    }
  }

  sum = sum / 16.0;

  shadowCoeff = projCoords.z - sum;
  shadowCoeff = 1.0 - (smoothstep(0.000, 0.085, shadowCoeff));

    // PCF 2
    // vec2 poissonDisk[4] = vec2[](vec2(-0.94201624f, -0.39906216f), vec2(0.94558609f, -0.76890725f), vec2(-0.094184101f, -0.92938870f), vec2(0.34495938f, 0.29387760f));
    // for(int i = 0; i < 4; i++) {
    //     if(texture(shadowMap, vec3(projCoords.xy + poissonDisk[i] / 700.0f, projCoords.z)) < projCoords.z) {
    //         shadowCoeff -= 0.2f;
    //     }
    // }

    // VSM
    // float distance = projCoords.z;
    // float mean = shadow_filter(shadowMap, projCoords, shadowMapSize);
    // float depth = texture(shadowMap, projCoords);
    // // note: normally depth_2 is stored in shadow map (in green channel)
    // float depth_2 = pow(depth, 2.0f);
    // float dx = dFdx(depth);
    // float dy = dFdy(depth);
    // depth_2 = depth_2 + 0.5f * (dx * dx + dy * dy);
    // float variance = depth_2 - pow(mean, 2.00f);
    // variance = max(variance, 0.005f);

    // float p = smoothstep(distance - 0.02f, distance, mean);
    // float d = distance - mean;

    // float p_max = linstep(0.2f, 1.0f, variance / (variance + d * d));
    // shadowCoeff = clamp(max(p, p_max), 0.0f, 1.0f);

    // shadowCoeff = 0 = shadow
    // shadowCoeff = 1 = no shadow
  return shadowCoeff;
    // To debug
    // if(projCoords.z > zShadowMap + EPSILON) {
    //     return 1.0f;
    // }

    // return PCF(projCoords, bias);
    // return PCFSampled(projCoords, bias);
    // return PCFSampled(projCoords, bias);

    // ivec2 shadowMapSize = textureSize(u_shadow_map, 0);

    // float xOffset = 1.0f / float(shadowMapSize.x);
    // float yOffset = 1.0f / float(shadowMapSize.y);
    // float factor = 0.0f;

    // for(int y = -1; y <= 1; y++) {
    //     for(int x = -1; x <= 1; x++) {
    //         vec2 offsets = vec2(float(x) * xOffset, float(y) * yOffset);
    //         vec3 uvc = vec3(projCoords.xy + offsets, projCoords.z + EPSILON);
    //         factor += texture(u_shadow_map, uvc);
    //     }
    // }

    // return (0.5f + (factor / 18.0f));

    // vec3 biased = vec3(projCoords.xy, projCoords.z - bias);

    // return texture(u_shadow_map, biased);

    // for(int i = 0; i < samples; i++) {
    //     // vec3 biased = vec3(projCoords.xy, projCoords.z - bias);
    //     vec3 biased = vec3(projCoords.xy + vec2(sampleOffsetDirections[i]) / shadowSpread, projCoords.z - bias);
    //     // vec3 biased = vec3(projCoords.xy + sampleOffsetDirections[i] * diskRadius, projCoords.z - bias);
    //     // vec3 biased = vec3(projCoords + sampleOffsetDirections[i] * diskRadius);
    //     // float litPercent = texture(u_shadow_map, vec3(projCoords.xy + adjacentPixels[i % 5] / shadowSpread, projCoords.z), 0.5f);

    //     // float litPercent = texture(u_shadow_map, vec3(projCoords.xy + vec2(sampleOffsetDirections[i % 5]) / shadowSpread, projCoords.z));
    //     float litPercent = texture(u_shadow_map, biased);
    //     visibility += litPercent;

    //     // float litPercent = texture(u_shadow_map, vec3(projCoords.xy + adjacentPixels[i % 5] / shadowSpread, projCoords.z), bias);
    //     // float litPercent = texture(u_shadow_map, vec3(projCoords.xy + adjacentPixels[i % 5] / shadowSpread, projCoords.z));
    //     // visibility *= max(litPercent, 0.8f);
    // }

    // return visibility / float(samples);
}

float ShadowCalculation(in sampler2DArrayShadow shadowMap, int index, vec4 fragPosLightSpace, float bias) {
  vec3 projCoords = fragPosLightSpace.xyz / fragPosLightSpace.w;
  projCoords = projCoords * 0.5 + 0.5;

  vec2 shadowMapSize = vec2(textureSize(shadowMap, 0));

  float shadowCoeff = 0.0;

    // BEGIN PCF
  float sum = 0.0;
  vec2 duv;

  for(float pcf_x = -1.5; pcf_x <= 1.5; pcf_x += 1.) {
    for(float pcf_y = -1.5; pcf_y <= 1.5; pcf_y += 1.) {
      duv = vec2(pcf_x / float(shadowMapSize.x), pcf_y / float(shadowMapSize.y));
      sum += shadow_filter(shadowMap, index, vec3(projCoords.xy, projCoords.z) + vec3(duv, 0.0), shadowMapSize);
    }
  }

  sum = sum / 16.0;

  shadowCoeff = projCoords.z - sum;
  shadowCoeff = 1.0 - (smoothstep(0.000, 0.085, shadowCoeff));
  return shadowCoeff;
}