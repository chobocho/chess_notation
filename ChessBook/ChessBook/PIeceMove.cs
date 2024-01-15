public interface PieceMove
{
    public bool KingMove(int to_x, int to_y, int from_x, int from_y);
    public bool QueenMove(int to_x, int to_y, int from_x, int from_y);
    public bool BishopMove(int to_x, int to_y, int from_x, int from_y);
    public bool KnightMove(int to_x, int to_y, int from_x, int from_y);
    public bool RookMove(int to_x, int to_y, int from_x, int from_y);
    public bool PawnMove(int to_x, int to_y, int from_x, int from_y);
}