using System.Runtime.CompilerServices;

class Piece
{
    public enum ChessPieceType
    {
        WhiteKing = '\u2654',
        BlackKing = '\u265a',
        WhiteQueen = '\u2655',
        BlackQueen = '\u265b',
        WhiteRook = '\u2656',
        BlackRook = '\u265c',
        WhiteBishop = '\u2657',
        BlackBishop = '\u265d',
        WhiteKnight = '\u2658',
        BlackKnight = '\u265e',
        WhitePawn = '\u2659',
        BlackPawn = '\u265f'
    }
    
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
        return (char)pieceType;
    }

    internal PieceType pieceType { get; set; }
    internal int x { get; set; }
    internal int y { get; set; }
    private bool isAlive;
    private bool isBlack;
}