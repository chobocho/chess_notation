using System.Xml;

class Piece
{
    public delegate bool Movable(int to_x, int to_y, int from_x, int from_y);

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
        this.color = color;
        this.pieceType = type;
        this.x = x;
        this.y = y;
        this.isAlive = true;
        this._movable = move;
    }

    public void setDead()
    {
        this.isAlive = false;
    }

    public void Move(int x, int y)
    {
        var color = (this.color == 0) ? 'B' : 'W';
        Console.WriteLine($"\n{color} {pieceType} : ({(char)(this.x + 'A')}, {this.y+1}) -> ({(char)(x + 'A')}, {y+1})\n");
        this.x = x;
        this.y = y;
    }

    public char getType()
    {
        //return isBlack ? BlackPiece[(char)pieceType] : WhitePiece[(char)pieceType];
        return (char)pieceType;
    }
    
    public static PieceType getPieceType(char type)
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

    public bool CanMove(int to_x, int to_y)
    {
        if (!isAlive) return false;
        return _movable != null ? _movable(to_x, to_y, this.x, this.y) : false;
    }

    internal int color { get; private set; }

    private PieceType pieceType { get; set; }
    internal int x { get; set; }
    internal int y { get; set; }
    private bool isAlive { get; set; }
    private Func<int, int, int, int, bool> _movable;
}