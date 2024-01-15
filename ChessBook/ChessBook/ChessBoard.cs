
public class ChessBoard
{
    private char[,] board = new char[8, 8];

    private Dictionary<Piece.PieceType, List<Piece>> pieceTable = new Dictionary<Piece.PieceType, List<Piece>>();
    private List<Piece> KingList = new List<Piece>();
    private List<Piece> QueenList = new List<Piece>();
    private List<Piece> BishopList = new List<Piece>();
    private List<Piece> KnightList = new List<Piece>();
    private List<Piece> RookList = new List<Piece>();
    private List<Piece> PawnList = new List<Piece>();
    private readonly PieceMove _pieceMove = new PieceMove();
    private int playerTurn = Piece.WHITE;


    private void GeneratePiece(int color, int y1, int y2)
    {
        KingList.Add(PieceFactory(color, Piece.PieceType.King, 4, y1, _pieceMove.KingMove));
        QueenList.Add( PieceFactory(color, Piece.PieceType.Queen, 3, y1, _pieceMove.QueenMove));
        
        BishopList.Add(PieceFactory(color, Piece.PieceType.Bishop, 2, y1, _pieceMove.BishopMove));
        BishopList.Add(PieceFactory(color, Piece.PieceType.Bishop, 5, y1, _pieceMove.BishopMove));
        
        KnightList.Add(PieceFactory(color, Piece.PieceType.Knight, 1, y1, _pieceMove.KnightMove));
        KnightList.Add(PieceFactory(color, Piece.PieceType.Knight, 6, y1, _pieceMove.KnightMove));
        
        RookList.Add(PieceFactory(color, Piece.PieceType.Rook, 0, y1, _pieceMove.RookMove));
        RookList.Add(PieceFactory(color, Piece.PieceType.Rook, 7, y1, _pieceMove.RookMove));
        
        for (int i = 8; i < 16; i++)
        {
            PawnList.Add(PieceFactory(color, Piece.PieceType.Pawn, i-8, y2, _pieceMove.PawnMove));
        }
    }
    
    public ChessBoard()
    {
        InitValues();

        GeneratePiece(Piece.WHITE, 0, 1);
        GeneratePiece(Piece.BLACK, 7, 6);

        UpdateBoard();
        _pieceMove = new PieceMove();
    }

    private void InitValues()
    {
        pieceTable[Piece.PieceType.King] = KingList;
        pieceTable[Piece.PieceType.Queen] = QueenList;
        pieceTable[Piece.PieceType.Bishop] = BishopList;
        pieceTable[Piece.PieceType.Knight] = KnightList;
        pieceTable[Piece.PieceType.Rook] = RookList;
        pieceTable[Piece.PieceType.Pawn] = PawnList;
    }

    private void UpdateBoard()
    {
        InitBoard();
        foreach (var entry in pieceTable)
        {
            List<Piece> pieces = entry.Value;
            foreach (Piece piece in pieces)
            {
                board[piece.y, piece.x] = (piece.color == Piece.BLACK) ? char.ToLower(piece.getType()) : piece.getType();
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

    Piece PieceFactory(int color, Piece.PieceType type, int x, int y, Func<int, int, int, int, int, bool> move)
    {
        Piece newPiece = new Piece();
        newPiece.Set(color, type, x, y, move);
        return newPiece;
    }
    
    public void Print()
    {
        Console.WriteLine("   A|B|C|D|E|F|G|H\n");
        for (var i = 7; i >= 0; --i)
        {
            Console.Write($"{i+1}  ");
            for (var j = 0; j < 8; j++)
            {
                if (char.IsLower(board[i,j]))
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
                Console.Write($"{char.ToUpper(board[i, j])} ");
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
        int y = move[2] - '0' - 1;

        Piece selectedPiece = FindPiece(playerTurn, type, x, y);
        playerTurn = 1 - playerTurn;
        if (selectedPiece != null)
        {
            selectedPiece.Move(x, y);
            UpdateBoard();
            return;
        }
        Console.WriteLine("Error: No Piece!\n");
    }

    private Piece FindPiece(int color, Piece.PieceType type, int x, int y)
    {
        List<Piece> pieces = pieceTable[type];
        foreach (Piece p in pieces)
        {
            if (p.color == color && p.CanMove(x, y))
            {
                return p;
            }
        }
        return null;
    }
}