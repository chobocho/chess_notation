namespace ChoboChessBoard;

public class ChessBoard
{
    private char[,] _board = new char[8, 8];

    private Dictionary<Piece.PieceType, List<Piece>> _pieceTable = new Dictionary<Piece.PieceType, List<Piece>>();
    private List<Piece> _KingList = new ();
    private List<Piece> _QueenList = new ();
    private List<Piece> _BishopList = new ();
    private List<Piece> _KnightList = new ();
    private List<Piece> _RookList = new ();
    private List<Piece> _PawnList = new ();
    private int _playerTurn = Piece.WHITE;

    public ChessBoard()
    {
        InitPieces();

        GeneratePiece(Piece.WHITE, 0, 1, new WhitePieceMove(_board));
        GeneratePiece(Piece.BLACK, 7, 6, new BlackPieceMove(_board));

        UpdateBoard();
    }

    private void GeneratePiece(int color, int y1, int y2, IPieceMove pieceMove)
    {
        _KingList.Add(PieceFactory(color, Piece.PieceType.King, 4, y1, pieceMove.KingMove));
        _QueenList.Add( PieceFactory(color, Piece.PieceType.Queen, 3, y1, pieceMove.QueenMove));
        
        _BishopList.Add(PieceFactory(color, Piece.PieceType.Bishop, 2, y1, pieceMove.BishopMove));
        _BishopList.Add(PieceFactory(color, Piece.PieceType.Bishop, 5, y1, pieceMove.BishopMove));
        
        _KnightList.Add(PieceFactory(color, Piece.PieceType.Knight, 1, y1, pieceMove.KnightMove));
        _KnightList.Add(PieceFactory(color, Piece.PieceType.Knight, 6, y1, pieceMove.KnightMove));
        
        _RookList.Add(PieceFactory(color, Piece.PieceType.Rook, 0, y1, pieceMove.RookMove));
        _RookList.Add(PieceFactory(color, Piece.PieceType.Rook, 7, y1, pieceMove.RookMove));
        
        for (int i = 8; i < 16; i++)
        {
            _PawnList.Add(PieceFactory(color, Piece.PieceType.Pawn, i-8, y2, pieceMove.PawnMove));
        }
    }
    
    private void InitPieces()
    {
        _pieceTable[Piece.PieceType.King] = _KingList;
        _pieceTable[Piece.PieceType.Queen] = _QueenList;
        _pieceTable[Piece.PieceType.Bishop] = _BishopList;
        _pieceTable[Piece.PieceType.Knight] = _KnightList;
        _pieceTable[Piece.PieceType.Rook] = _RookList;
        _pieceTable[Piece.PieceType.Pawn] = _PawnList;
    }

    private void UpdateBoard()
    {
        InitBoard();
        foreach (var entry in _pieceTable)
        {
            var pieces = entry.Value;
            foreach (var piece in pieces)
            {
                _board[piece._y, piece._x] = piece._color == Piece.BLACK ? 
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
                _board[i, j] = ' ';
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
        Console.ForegroundColor = char.IsLower(_board[i,j]) ? ConsoleColor.Black : ConsoleColor.Red;
        Console.BackgroundColor = ((i + j) & 0x1) == 1 ? ConsoleColor.Gray : ConsoleColor.White;
        Console.Write($"{char.ToUpper(_board[i, j])} ");
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
        var selectedPiece = FindPiece(_playerTurn, pieceType, x, y);
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
        var pieceType = Piece.GetPieceType(move[0]);
        var destinationColumn = move[1] - 'a';
        var destinationRow = move[2] - '1'; // assuming chess board rows start from 1
        return (pieceType, destinationColumn, destinationRow);
    }
    
    
    private void TogglePlayerTurn()
    {
        _playerTurn = 1 - _playerTurn;
    }
    
    private Piece FindPiece(int color, Piece.PieceType type, int toX, int toY, int fromX, int fromY, bool isCatched)
    {
        return _pieceTable[type].FirstOrDefault(p => p._color == color && p.CanMove(toX, toY));
    }
    
    private Piece FindPiece(int color, Piece.PieceType type, int x, int y)
    {
        return _pieceTable[type].FirstOrDefault(p => p._color == color && p.CanMove(x, y));
    }
}