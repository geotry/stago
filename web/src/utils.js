export const loadImageBitmap = async (url) => {
  const res = await fetch(url);
  const blob = await res.blob();
  return await createImageBitmap(blob, { colorSpaceConversion: 'none' });
};