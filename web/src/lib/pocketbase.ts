import PocketBase, { type RecordModel } from 'pocketbase';
import { env } from '$env/dynamic/public';

const PUBLIC_API_URL = env.PUBLIC_API_URL;

console.log('API URL:', PUBLIC_API_URL);
const pb = new PocketBase(PUBLIC_API_URL); // Replace with your PocketBase URL
pb.autoCancellation(false);

export async function subscribeToAuctionUpdate(recordId: string, callback: (record: RecordModel) => void) {
  await pb.collection('auctions').subscribe(recordId, (e) => {
    console.log('Auction updated:', e);
    callback(e.record);
  });
}

export async function unsubscribeFromAuctionUpdates() {
  pb.collection('auctions').unsubscribe();
}

export async function subscribeToUserUpdate(userId: string, callback: (record: RecordModel) => void) {
  await pb.collection('users').subscribe(userId, (e) => {
    console.log('User updated:', e);
    callback(e.record);
  });
}
export async function unsubscribeFromUserUpdates() {
  await pb.collection('users').unsubscribe();
}
export function getUser() {
  return pb.authStore.record as RecordModel | null;
}

export default pb;
