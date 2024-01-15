
public class ChessBoard
{
    private char[,] board = new char[8, 8];

    private Dictionary<Piece.PieceType, List<Piece>> pieceTable = new Dictionary<Piece.PieceType, List<Piece>>();
    private List<Piece> KingList = new ();
    private List<Piece> QueenList = new ();
    private List<Piece> BishopList = new ();
    private List<Piece> KnightList = new ();
    private List<Piece> RookList = new ();
    private List<Piece> PawnList = new ();
    private readonly PieceMove _blackPieceMove = new BlackPieceMove();
    private readonly PieceMove _whitePieceMove = new WhitePieceMove();
    private int playerTurn = Piece.WHITE;

    public ChessBoard()
    {
        InitPieces();

        GeneratePiece(Piece.WHITE, 0, 1, _whitePieceMove);
        GeneratePiece(Piece.BLACK, 7, 6, _blackPieceMove);

        UpdateBoard();
    }

    private void GeneratePiece(int color, int y1, int y2, PieceMove pieceMove)
    {
        KingList.Add(PieceFactory(color, Piece.PieceType.King, 4, y1, pieceMove.KingMove));
        QueenList.Add( PieceFactory(color, Piece.PieceType.Queen, 3, y1, pieceMove.QueenMove));
        
        BishopList.Add(PieceFactory(color, Piece.PieceType.Bishop, 2, y1, pieceMove.BishopMove));
        BishopList.Add(PieceFactory(color, Piece.PieceType.Bishop, 5, y1, pieceMove.BishopMove));
        
        KnightList.Add(PieceFactory(color, Piece.PieceType.Knight, 1, y1, pieceMove.KnightMove));
        KnightList.Add(PieceFactory(color, Piece.PieceType.Knight, 6, y1, pieceMove.KnightMove));
        
        RookList.Add(PieceFactory(color, Piece.PieceType.Rook, 0, y1, pieceMove.RookMove));
        RookList.Add(PieceFactory(color, Piece.PieceType.Rook, 7, y1, pieceMove.RookMove));
        
        for (int i = 8; i < 16; i++)
        {
            PawnList.Add(PieceFactory(color, Piece.PieceType.Pawn, i-8, y2, pieceMove.PawnMove));
        }
    }
    
    private void InitPieces()
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
            var pieces = entry.Value;
            foreach (var piece in pieces)
            {
                board[piece.y, piece.x] = (piece.color == Piece.BLACK) ? 
                    char.ToLower(piece.getType()) : piece.getType();
            }
        }
    }

    private void InitBoard()
    {
        for (var i = 0; i < 8; i++)
        {
            for (var j = 0; j < 8; j++)
            {
                board[i, j] = ' ';
            }
        }
    }

    Piece PieceFactory(int color, Piece.PieceType type, int x, int y, Func<int, int, int, int, bool> move)
    {
        var newPiece = new Piece();
        newPiece.Set(color, type, x, y, move);
        return newPiece;
    }
    
    public void Print()
    {
        const string boardHeaders = "   A|B|C|D|E|F|G|H\n";
        Console.WriteLine(boardHeaders);
        for (var i = 7; i >= 0; --i)
        {
            Console.Write($"{i + 1}  ");
            for (var j = 0; j < 8; j++)
            {
                PrintSquare(i, j);
            }
            Console.WriteLine("");
        }
    }
    
    private void PrintSquare(int i, int j)
    {
        Console.ForegroundColor = char.IsLower(board[i,j]) ? ConsoleColor.Black : ConsoleColor.Red;
        Console.BackgroundColor = ((i + j) & 0x1) == 1 ? ConsoleColor.Gray : ConsoleColor.White;
        Console.Write($"{char.ToUpper(board[i, j])} ");
        Console.BackgroundColor = ConsoleColor.Black;
        Console.ForegroundColor = ConsoleColor.Gray;
    }

    public void MoveBack()
    {
       // ToDo
       // Implement
    }

    public void MoveNext(string move)
    {
        var (pieceType, x, y) = ParseMove(move);
        var selectedPiece = FindPiece(playerTurn, pieceType, x, y);
        TogglePlayerTurn();
        
        if (selectedPiece == null)
        {
            Console.WriteLine("Error: No Piece!\n");
            return;
        }
        
        selectedPiece.Move(x, y);
        UpdateBoard();
    }

    private (Piece.PieceType, int, int) ParseMove(string move)
    {
        var pieceType = Piece.getPieceType(move[0]);
        var destinationColumn = move[1] - 'a';
        var destinationRow = move[2] - '1'; // assuming chess board rows start from 1
        return (pieceType, destinationColumn, destinationRow);
    }
    
    
    private void TogglePlayerTurn()
    {
        playerTurn = 1 - playerTurn;
    }
    
    private Piece FindPiece(int color, Piece.PieceType type, int x, int y)
    {
        return pieceTable[type].FirstOrDefault(p => p.color == color && p.CanMove(x, y));
    }
}