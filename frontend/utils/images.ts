// Hosted requests must leave room for multipart fields below Vercel's body cap.
export const MAX_IMAGE_MB = process.env.NEXT_PUBLIC_MAX_IMAGE_MB === '4' ? 4 : 5;

export function imageSizeError(file: File): string | null {
  return file.size > MAX_IMAGE_MB * 1024 * 1024
    ? `Image file size exceeds ${MAX_IMAGE_MB}MB limit.`
    : null;
}
