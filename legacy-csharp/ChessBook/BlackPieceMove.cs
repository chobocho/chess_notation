namespace ChoboChessBoard;

public class BlackPieceMove : IPieceMove
{
    private char[,] _board;

    public BlackPieceMove(char[,] board)
    {
        _board = board;
    }

    public bool KingMove(int toX, int toY, int fromX, int fromY)
    {
        return false;
    }

    public bool QueenMove(int toX, int toY, int fromX, int fromY)
    {
        return (toY == fromY && toX != fromX) ||
               (toY != fromY && toX == fromX) ||
               (toY - fromY == toX - fromX);
    }

    public bool BishopMove(int toX, int toY, int fromX, int fromY)
    {
        return (toY - fromY == toX - fromX);
    }

    public bool KnightMove(int toX, int toY, int fromX, int fromY)
    {
        return false;
    }

    public bool RookMove(int toX, int toY, int fromX, int fromY)
    {
        return (toY == fromY && toX != fromX) ||
               (toY != fromY && toX == fromX);
    }

    public bool PawnMove(int toX, int toY, int fromX, int fromY)
    {
        // Console.WriteLine($"{color} PawnMove : ({(char)(from_x + 'A')}, {from_y}) -> ({(char)(to_x + 'A')}, {to_y})");
        if (toY >= fromY) return false;
        if (fromX == toX && fromY - 1 == toY) return true;
        if (fromX == toX && (fromY - 2 == toY) && fromY == 6) return true;

        return false;
    }
}