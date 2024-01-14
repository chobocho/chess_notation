using System.Runtime.CompilerServices;

class Piece
{
    // private Dictionary<char, char> BlackPiece = new Dictionary<char, char>()
    // {
    //     {'K', '\u265a'},
    //     {'Q', '\u265b'},
    //     {'B', '\u265d'},
    //     {'N', '\u265e'},
    //     {'R', '\u265c'},
    //     {'P', '\u265f'}
    // };
    //
    // private Dictionary<char, char> WhitePiece = new Dictionary<char, char>()
    // {
    //     {'K', '\u2654'},
    //     {'Q', '\u2655'},
    //     {'B', '\u2657'},
    //     {'N', '\u2658'},
    //     {'R', '\u2656'},
    //     {'P', '\u2659'}
    // };
    //

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

    public void Set(bool black, PieceType type, int x, int y)
    {
        this.isBlack = black;
        this.pieceType = type;
        this.x = x;
        this.y = y;
        this.isAlive = true;
    }

    public void setDead()
    {
        this.isAlive = false;
    }

    public void Move(int x, int y)
    {
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
    
    
    internal PieceType pieceType { get; set; }
    internal int x { get; set; }
    internal int y { get; set; }
    private bool isAlive;
    private bool isBlack;
}