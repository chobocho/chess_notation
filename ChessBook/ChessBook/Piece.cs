namespace ChoboChessBoard;

class Piece
{
    public const int BLACK = 0;
    public const int WHITE = 1;
    
    public enum PieceType
    {
        King = 'K',
        Queen = 'Q',
        Rook = 'R',
        Bishop = 'B',
        Knight = 'N',
        Pawn = 'P'
    }
    
    public Piece()
    {

    }

    public void Set(int color, PieceType type, int x, int y, Func<int, int, int, int, bool> move)
    {
        _color = color;
        _pieceType = type;
        _x = x;
        _y = y;
        _isAlive = true;
        _movable = move;
    }

    public void setDead()
    {
        _isAlive = false;
    }

    public void Move(int x, int y)
    {
        var color = _color == 0 ? 'B' : 'W';
        Console.WriteLine($"\n{color} {_pieceType} : ({(char)(_x + 'A')}, {_y+1}) -> ({(char)(x + 'A')}, {y+1})\n");
        _x = x;
        _y = y;
    }

    public char getType()
    {
        //return isBlack ? BlackPiece[(char)pieceType] : WhitePiece[(char)pieceType];
        return (char)_pieceType;
    }
    
    public static PieceType GetPieceType(char type)
    {
        switch (type) 
        {
            case 'K':
                return PieceType.King;
            case 'Q':
                return PieceType.Queen;
            case 'R':
                return PieceType.Rook;
            case 'B':
                return PieceType.Bishop;
            case 'N':
                return PieceType.Knight;
            case 'P':
                return PieceType.Pawn;
            default:
                Console.WriteLine("Never come to here!");
                return PieceType.Pawn;
        }
    }

    public bool CanMove(int toX, int toY)
    {
        if (!_isAlive) return false;
        return _movable != null ? _movable(toX, toY, _x, _y) : false;
    }

    internal int _color { get; private set; }

    private PieceType _pieceType { get; set; }
    internal int _x { get; set; }
    internal int _y { get; set; }
    private bool _isAlive { get; set; }
    private Func<int, int, int, int, bool> _movable;
}