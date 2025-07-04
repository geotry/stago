/**
 * Fetch the fragment and vertex shader text from external files.
 *
 * @param shaderName
 * @returns {Promise<{vertexShaderSource: string | null, fragmentShaderSource: string | null}>}
 */
export async function loadShader(shaderName) {
  const results = {
    vertexShaderSource: null,
    fragmentShaderSource: null,
  };

  const vertexShaderPath = `/shaders/${shaderName}/vertex.glsl`;
  const fragmentShaderPath = `/shaders/${shaderName}/fragment.glsl`;

  let errors = [];
  await Promise.all([
    fetch(vertexShaderPath)
      .catch((e) => {
        errors.push(e);
      })
      .then(async (response) => {
        if (response.status === 200) {
          results.vertexShaderSource = await response.text();
        } else {
          errors.push(
            `Non-200 response for ${vertexShaderPath}.  ${response.status}:  ${response.statusText}`
          );
        }
      }),

    fetch(fragmentShaderPath)
      .catch((e) => errors.push(e))
      .then(async (response) => {
        if (response.status === 200) {
          results.fragmentShaderSource = await response.text();
        } else {
          errors.push(
            `Non-200 response for ${fragmentShaderPath}.  ${response.status}:  ${response.statusText}`
          );
        }
      }),
  ]);

  if (errors.length !== 0) {
    throw new Error(
      `Failed to fetch shader(s):\n${JSON.stringify(errors, (key, value) => {
        if (value?.constructor.name === 'Error') {
          return {
            name: value.name,
            message: value.message,
            stack: value.stack,
            cause: value.cause,
          };
        }
        return value;
      }, 2)}`
    );
  }
  return results;
}