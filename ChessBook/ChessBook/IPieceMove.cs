namespace ChoboChessBoard;

public interface IPieceMove
{
    public bool KingMove(int toX, int toY, int fromX, int fromY);
    public bool QueenMove(int toX, int toY, int fromX, int fromY);
    public bool BishopMove(int toX, int toY, int fromX, int fromY);
    public bool KnightMove(int toX, int toY, int fromX, int fromY);
    public bool RookMove(int toX, int toY, int fromX, int fromY);
    public bool PawnMove(int toX, int toY, int fromX, int fromY);
}