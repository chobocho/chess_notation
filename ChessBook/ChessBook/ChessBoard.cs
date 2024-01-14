public class ChessBoard
{
    private const int BLACK = 0;
    private const int WHITE = 1;

    private char[,] board = new char[8, 8];
    private Piece[,] piece = new Piece[2, 16];

    private Dictionary<Piece.PieceType, List<Piece>> pieceTable = new Dictionary<Piece.PieceType, List<Piece>>();
    private List<Piece> KingList = new List<Piece>();
    private List<Piece> QueenList = new List<Piece>();
    private List<Piece> BishopList = new List<Piece>();
    private List<Piece> KnightList = new List<Piece>();
    private List<Piece> RookList = new List<Piece>();
    private List<Piece> PawnList = new List<Piece>();
    
    private void GeneratePiece(int color, int y1, int y2)
    {
        piece[color, 0] = PieceFactory(color, Piece.PieceType.King, 4, y1);
        KingList.Add(piece[color, 0]);
        piece[color, 1] = PieceFactory(color, Piece.PieceType.Queen, 3, y1);
        QueenList.Add(piece[color, 1]);
        
        piece[color, 2] = PieceFactory(color, Piece.PieceType.Bishop, 2, y1);
        BishopList.Add(piece[color, 2]);
        piece[color, 3] = PieceFactory(color, Piece.PieceType.Bishop, 5, y1);
        BishopList.Add(piece[color, 3]);
        
        piece[color, 4] = PieceFactory(color, Piece.PieceType.Knight, 1, y1);
        KnightList.Add(piece[color, 4]);
        piece[color, 5] = PieceFactory(color, Piece.PieceType.Knight, 6, y1);
        KnightList.Add(piece[color, 5]);
        
        piece[color, 6] = PieceFactory(color, Piece.PieceType.Rook, 0, y1);
        RookList.Add(piece[color, 6]);
        piece[color, 7] = PieceFactory(color, Piece.PieceType.Rook, 7, y1);
        RookList.Add(piece[color, 7]);
        
        for (int i = 8; i < 16; i++)
        {
            piece[color, i] = PieceFactory(color, Piece.PieceType.Pawn, i-8, y2);
            PawnList.Add(piece[color, i]);
        }
    }
    
    public ChessBoard()
    {
        InitValues();

        GeneratePiece(BLACK, 0, 1);
        GeneratePiece(WHITE, 7, 6);

        UpdateBoard();
    }

    private void InitValues()
    {
        pieceTable[Piece.PieceType.King] = KingList;
        pieceTable[Piece.PieceType.Queen] = QueenList;
        pieceTable[Piece.PieceType.Bishop] = BishopList;
        pieceTable[Piece.PieceType.Knight] = KingList;
        pieceTable[Piece.PieceType.Rook] = RookList;
        pieceTable[Piece.PieceType.Pawn] = PawnList;
    }

    private void UpdateBoard()
    {
        InitBoard();
        for (int color = 0; color <= 1; color++)
        {
            for (int i = 0; i < 16; i++)
            {
                int x = piece[color, i].x;
                int y = piece[color, i].y;
                board[y, x] = piece[color, i].getType();
            }
        }
    }

    private void InitBoard()
    {
        for (var i = 0; i < 8; i++)
        {
            for (var j = 0; j < 8; j++)
            {
                // board[i, j] = ((i + j) & 0x1) == 0x1 ? '\u25a0' : '\u25a1';
                board[i, j] = ' ';
            }
        }
    }

    Piece PieceFactory(int color, Piece.PieceType type, int x, int y)
    {
        Piece newPiece = new Piece(); 
        newPiece.Set(color == BLACK, type, x, y);
        return newPiece;
    }
    
    public void Print()
    {
        Console.WriteLine("   A|B|C|D|E|F|G|H\n");
        for (var i = 0; i < 8; i++)
        {
            Console.Write($"{8-i}  ");
            for (var j = 0; j < 8; j++)
            {
                if (i < 2)
                {
                    Console.ForegroundColor = ConsoleColor.Black;
                }
                else
                {
                    Console.ForegroundColor = ConsoleColor.Red;
                }

                if (((i + j) & 0x1) == 0x1)
                {
                    Console.BackgroundColor = ConsoleColor.Gray;
                }
                else
                {
                    Console.BackgroundColor = ConsoleColor.White;
                }
                Console.Write($"{board[i, j]} ");
                Console.BackgroundColor = ConsoleColor.Black;
                Console.ForegroundColor = ConsoleColor.Gray;
            }
            Console.WriteLine("");
        }
    }

    public void MoveBack()
    {
       // ToDo
       // Implement
    }

    public void MoveNext(string move)
    {
        // ToDo
        Piece.PieceType type = Piece.getPieceType(move[0]);
        int x = move[1] - 'a';
        int y = move[2] - '0';

        Piece selectedPiece = FindPiece(type, x, y);
        selectedPiece.Move(x, y);
    }

    private Piece FindPiece(Piece.PieceType type, int x, int y)
    {
        
    }
}