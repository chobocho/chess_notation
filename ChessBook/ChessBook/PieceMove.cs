public class PieceMove
{
    public PieceMove()
    {
    }

    public bool KingMove(int color, int to_x, int to_y, int from_x, int from_y)
    {
        return false;
    }

    public bool QueenMove(int color, int to_x, int to_y, int from_x, int from_y)
    {
        return false;
    }

    public bool BishopMove(int color, int to_x, int to_y, int from_x, int from_y)
    {
        return false;
    }

    public bool KnightMove(int color, int to_x, int to_y, int from_x, int from_y)
    {
        return false;
    }

    public bool RookMove(int color, int to_x, int to_y, int from_x, int from_y)
    {
        return false;
    }

    public bool PawnMove(int color, int to_x, int to_y, int from_x, int from_y)
    {
        // Console.WriteLine($"{color} PawnMove : ({(char)(from_x + 'A')}, {from_y}) -> ({(char)(to_x + 'A')}, {to_y})");
        if (color == Piece.BLACK)
        {
            if (to_y >= from_y) return false;
            if (from_x == to_x && from_y - 1 == to_y) return true;
            if (from_x == to_x && (from_y - 2 == to_y) && from_y == 6) return true;
        }
        
        if (to_y <= from_y) return false;
        if (from_x == to_x && from_y + 1 == to_y) return true;
        if (from_x == to_x && (from_y + 2) == to_y && from_y == 1) return true;

        return false;
    }
}