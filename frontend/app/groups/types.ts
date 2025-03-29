export interface Group {
  id: number;
  title: string;
  description: string;
  creator_id: number;
  created_at: string;
  member_count: number;
  role?: string;
  status?: string;
}
